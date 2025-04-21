package broadcaster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	zena "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"

	authTypes "github.com/zenanetwork/iris/auth/types"
	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/helper"

	"github.com/tendermint/tendermint/libs/log"

	hmTypes "github.com/zenanetwork/iris/types"
)

// TxBroadcaster uses to broadcast transaction to each chain
type TxBroadcaster struct {
	logger log.Logger

	CliCtx cliContext.CLIContext

	irisMutex  sync.Mutex
	maticMutex sync.Mutex

	lastSeqNo uint64
	accNum    uint64
}

// NewTxBroadcaster creates new broadcaster
func NewTxBroadcaster(cdc *codec.Codec) *TxBroadcaster {
	cliCtx := cliContext.NewCLIContext().WithCodec(cdc)
	cliCtx.BroadcastMode = client.BroadcastSync
	cliCtx.TrustNode = true

	// current address
	address := hmTypes.BytesToIrisAddress(helper.GetAddress())

	account, err := util.GetAccount(cliCtx, address)
	if err != nil {
		panic("Error connecting to rest-server, please start server before bridge.")
	}

	return &TxBroadcaster{
		logger:    util.Logger().With("module", "txBroadcaster"),
		CliCtx:    cliCtx,
		lastSeqNo: account.GetSequence(),
		accNum:    account.GetAccountNumber(),
	}
}

// BroadcastToIris broadcast to iris
func (tb *TxBroadcaster) BroadcastToIris(msg sdk.Msg, event interface{}, testOpts ...*helper.TestOpts) (sdk.TxResponse, error) {
	tb.irisMutex.Lock()
	defer tb.irisMutex.Unlock()
	defer util.LogElapsedTimeForStateSyncedEvent(event, "BroadcastToIris", time.Now())

	// tx encoder
	txEncoder := helper.GetTxEncoder(tb.CliCtx.Codec)
	// chain id
	chainID := helper.GetGenesisDoc().ChainID

	// get account number and sequence
	txBldr := authTypes.NewTxBuilderFromCLI().
		WithTxEncoder(txEncoder).
		WithAccountNumber(tb.accNum).
		WithSequence(tb.lastSeqNo).
		WithChainID(chainID)

	txResponse, err := helper.BuildAndBroadcastMsgs(tb.CliCtx, txBldr, []sdk.Msg{msg}, testOpts...)
	if err != nil || txResponse.Code != uint32(sdk.CodeOK) {
		tb.logger.Error("Error while broadcasting the iris transaction", "error", err, "txResponse", txResponse.Code)

		// current address
		address := hmTypes.BytesToIrisAddress(helper.GetAddress())

		// fetch from APIs
		account, errAcc := util.GetAccount(tb.CliCtx, address)
		if errAcc != nil {
			tb.logger.Error("Error fetching account from rest-api", "url", helper.GetIrisServerEndpoint(fmt.Sprintf(util.AccountDetailsURL, helper.GetAddress())))
			return txResponse, errAcc
		}

		// update seqNo for safety
		tb.lastSeqNo = account.GetSequence()

		return txResponse, err
	}

	txHash := txResponse.TxHash

	tb.logger.Info("Tx sent on iris", "txHash", txHash, "accSeq", tb.lastSeqNo, "accNum", tb.accNum)
	tb.logger.Debug("Tx successful on iris", "txResponse", txResponse)
	// increment account sequence
	tb.lastSeqNo += 1

	return txResponse, nil
}

// BroadcastToMatic broadcast to matic
func (tb *TxBroadcaster) BroadcastToMatic(msg zena.CallMsg) error {
	tb.maticMutex.Lock()
	defer tb.maticMutex.Unlock()

	// get matic client
	maticClient := helper.GetMaticClient()

	// get auth
	auth, err := helper.GenerateAuthObj(maticClient, *msg.To, msg.Data)

	if err != nil {
		tb.logger.Error("Error generating auth object", "error", err)
		return err
	}

	// Create the transaction, sign it and schedule it for execution
	rawTx := types.NewTransaction(auth.Nonce.Uint64(), *msg.To, msg.Value, auth.GasLimit, auth.GasPrice, msg.Data)

	// signer
	signedTx, err := auth.Signer(auth.From, rawTx)
	if err != nil {
		tb.logger.Error("Error signing the transaction", "error", err)
		return err
	}

	tb.logger.Info("Sending transaction to zena", "txHash", signedTx.Hash())

	// create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), helper.GetConfig().ZenaRPCTimeout)
	defer cancel()

	// broadcast transaction
	if err := maticClient.SendTransaction(ctx, signedTx); err != nil {
		tb.logger.Error("Error while broadcasting the transaction to maticchain", "error", err)
		return err
	}

	return nil
}

// BroadcastToRootchain broadcast to rootchain
func (tb *TxBroadcaster) BroadcastToRootchain() {}
