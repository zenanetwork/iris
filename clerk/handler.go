package clerk

import (
	"encoding/hex"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/zenanetwork/iris/clerk/types"
	"github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
)

// NewHandler creates new handler for handling messages for clerk module
func NewHandler(k Keeper, contractCaller helper.IContractCaller) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgEventRecord:
			return handleMsgEventRecord(ctx, msg, k, contractCaller)
		default:
			return sdk.ErrTxDecode("Invalid message in clerk module").Result()
		}
	}
}

func handleMsgEventRecord(ctx sdk.Context, msg types.MsgEventRecord, k Keeper, _ helper.IContractCaller) sdk.Result {
	k.Logger(ctx).Debug("✅ Validating clerk msg",
		"id", msg.ID,
		"contract", msg.ContractAddress,
		"data", hex.EncodeToString(msg.Data),
		"txHash", hmTypes.BytesToIrisHash(msg.TxHash.Bytes()),
		"logIndex", msg.LogIndex,
		"blockNumber", msg.BlockNumber,
	)

	// check if event record exists
	if exists := k.HasEventRecord(ctx, msg.ID); exists {
		return types.ErrEventRecordAlreadySynced(k.Codespace()).Result()
	}

	// chainManager params
	params := k.chainKeeper.GetParams(ctx)
	chainParams := params.ChainParams

	// check chain id
	if chainParams.ZenaChainID != msg.ChainID {
		k.Logger(ctx).Error("Invalid Zena chain id", "msgChainID", msg.ChainID, "zenaChainId", chainParams.ZenaChainID)
		return common.ErrInvalidZenaChainID(k.Codespace()).Result()
	}

	// sequence id
	blockNumber := new(big.Int).SetUint64(msg.BlockNumber)
	sequence := new(big.Int).Mul(blockNumber, big.NewInt(hmTypes.DefaultLogIndexUnit))
	sequence.Add(sequence, new(big.Int).SetUint64(msg.LogIndex))

	// check if incoming tx is older
	if k.HasRecordSequence(ctx, sequence.String()) {
		k.Logger(ctx).Error("Older invalid tx found", "Sequence", sequence.String())
		return common.ErrOldTx(k.Codespace()).Result()
	}

	// add events
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRecord,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyRecordID, strconv.FormatUint(msg.ID, 10)),
			sdk.NewAttribute(types.AttributeKeyRecordContract, msg.ContractAddress.String()),
			sdk.NewAttribute(types.AttributeKeyRecordTxHash, msg.TxHash.String()),
			sdk.NewAttribute(types.AttributeKeyRecordTxLogIndex, strconv.FormatUint(msg.LogIndex, 10)),
		),
	})

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
