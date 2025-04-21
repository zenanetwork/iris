package broadcaster

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"

	cosmosCtx "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/mock/gomock"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/zenanetwork/iris/app"
	authTypes "github.com/zenanetwork/iris/auth/types"
	checkpointTypes "github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/helper"
	helperMocks "github.com/zenanetwork/iris/helper/mocks"
	hmTypes "github.com/zenanetwork/iris/types"
	zenaTypes "github.com/zenanetwork/iris/zena/types"
)

var (
	privKey                = secp256k1.GenPrivKey()
	pubKey                 = privKey.PubKey()
	address                = pubKey.Address()
	irisAddress            = hmTypes.BytesToIrisAddress([]byte(address))
	defaultBalance         = sdk.NewIntFromBigInt(big.NewInt(10).Exp(big.NewInt(10), big.NewInt(18), nil))
	testChainId            = "testChainId"
	dummyTenderMintNodeUrl = "http://localhost:26657"
	dummyIrisServerUrl     = "https://dummy-iris-api-testnet.polygon.technology"
	getAccountUrl          = dummyIrisServerUrl + "/auth/accounts/" + address.String()
	getAccountResponse     = fmt.Sprintf(`
	{
		"height": "11384869",
		"result": {
		  "type": "auth/Account",
		  "value": {
			"address": "0x%s",
			"coins": [
			  {
				"denom": "matic",
				"amount": "10000000000000000000"
			  }
			],
			"public_key": {
				"type": "tendermint/PubKeySecp256k1",
				"value": "BE/WIL+R3P+8YlGBfxqPdb+jWlWdAiocPOBYNXoXqYOlQ0+QiJudDIMLhDqovssOvS9REFaUYn6pXE0YGD3nb5k="
			  },
			"account_number": "0",
			"sequence": "0"
		  }
		}
	  }
	  `, address.String())

	getAccountUpdatedResponse = fmt.Sprintf(`
	{
		"height": "11384869",
		"result": {
		  "type": "auth/Account",
		  "value": {
			"address": "0x%s",
			"coins": [
			  {
				"denom": "matic",
				"amount": "10000000000000000000"
			  }
			],
			"public_key": {
				"type": "tendermint/PubKeySecp256k1",
				"value": "BE/WIL+R3P+8YlGBfxqPdb+jWlWdAiocPOBYNXoXqYOlQ0+QiJudDIMLhDqovssOvS9REFaUYn6pXE0YGD3nb5k="
			  },
			"account_number": "0",
			"sequence": "1"
		  }
		}
	  }
	  `, address.String())

	msgs = []sdk.Msg{
		checkpointTypes.NewMsgCheckpointBlock(
			irisAddress,
			0,
			63,
			hmTypes.HexToIrisHash("0x5bd83f679c8ce7c48d6fa52ce41532fcacfbbd99d5dab415585f397bf44a0b6e"),
			hmTypes.HexToIrisHash("0xd10b5c16c25efe0b0f5b3d75038834223934ae8c2ec2b63a62bbe42aa21e2d2d"),
			"zenaChainID",
		),
		checkpointTypes.NewMsgMilestoneBlock(
			irisAddress,
			0,
			63,
			hmTypes.HexToIrisHash("0x5bd83f679c8ce7c48d6fa52ce41532fcacfbbd99d5dab415585f397bf44a0b6e"),
			"testZenaChainID",
			"testMilestoneID",
		),
		checkpointTypes.NewMsgMilestoneTimeout(
			irisAddress,
		),
		zenaTypes.NewMsgProposeSpan(
			1,
			irisAddress,
			0,
			63,
			"testZenaChainID",
			common.Hash(hmTypes.BytesToIrisHash([]byte("randseed"))),
		),
	}
)

//nolint:tparallel
func TestBroadcastToIris(t *testing.T) {
	t.Parallel()

	viper.Set(helper.TendermintNodeFlag, dummyTenderMintNodeUrl)
	viper.Set("log_level", "info")

	configuration := helper.GetDefaultIrisConfig()
	configuration.TendermintRPCUrl = dummyTenderMintNodeUrl
	configuration.IrisServerURL = dummyIrisServerUrl
	helper.SetTestConfig(configuration)
	helper.SetTestPrivPubKey(privKey)

	mockCtrl := prepareMockData(t)
	defer mockCtrl.Finish()

	testOpts := helper.NewTestOpts(nil, testChainId)
	irisApp, sdkCtx, _ := createTestApp(false, testOpts)
	testOpts.SetApplication(irisApp)
	txBroadcaster := NewTxBroadcaster(irisApp.Codec())
	txBroadcaster.CliCtx.Simulate = true

	testCases := []struct {
		name       string
		msg        sdk.Msg
		op         func(*app.IrisApp) error
		expResCode uint32
		expErr     bool
		tearDown   func(*app.IrisApp) error
	}{
		{
			name: "successful broadcast",
			msg:  msgs[0],

			op:         nil,
			expResCode: 0,
			expErr:     false,
		},
		{
			name: "failed broadcast (insufficient funds for fees)",
			msg:  msgs[1],
			op: func(hApp *app.IrisApp) error {
				acc := hApp.AccountKeeper.GetAccount(sdkCtx, irisAddress)
				// reduce account balance to 0
				if err := acc.SetCoins(sdk.Coins{}); err != nil {
					return err
				}
				hApp.AccountKeeper.SetAccount(sdkCtx, acc)
				return nil
			},
			expResCode: 5,
			expErr:     true,
			tearDown: func(hApp *app.IrisApp) error {
				acc := hApp.AccountKeeper.GetAccount(sdkCtx, irisAddress)
				// reset account balance
				if err := acc.SetCoins(sdk.Coins{sdk.Coin{Denom: authTypes.FeeToken, Amount: defaultBalance}}); err != nil {
					return err
				}
				hApp.AccountKeeper.SetAccount(sdkCtx, acc)
				return nil
			},
		},
		{
			name: "failed broadcast (invalid sequence number)",
			msg:  msgs[2],
			op: func(hApp *app.IrisApp) error {
				acc := hApp.AccountKeeper.GetAccount(sdkCtx, irisAddress)
				txBroadcaster.lastSeqNo = acc.GetSequence() + 1
				return nil
			},
			expResCode: 4,
			expErr:     true,
		},
	}

	//nolint:paralleltest
	for _, tc := range testCases {
		if tc.expErr {
			updateMockData(t)
		}
		t.Run(tc.name, func(t *testing.T) {
			if tc.op != nil {
				err := tc.op(irisApp)
				require.NoError(t, err)
			}
			txRes, err := txBroadcaster.BroadcastToIris(tc.msg, nil, testOpts)
			require.NoError(t, err)
			require.Equal(t, tc.expResCode, txRes.Code)
			accSeq, err := irisApp.AccountKeeper.GetSequence(sdkCtx, irisAddress)
			require.NoError(t, err)
			require.Equal(t, txBroadcaster.lastSeqNo, accSeq)

			if tc.tearDown != nil {
				err := tc.tearDown(irisApp)
				require.NoError(t, err)
			}
		})
	}
}

func createTestApp(isCheckTx bool, testOpts *helper.TestOpts) (*app.IrisApp, sdk.Context, cosmosCtx.CLIContext) {
	hApp := app.Setup(isCheckTx, testOpts)
	ctx := hApp.BaseApp.NewContext(true, abci.Header{ChainID: testOpts.GetChainId()})
	hApp.BankKeeper.SetSendEnabled(ctx, true)
	hApp.AccountKeeper.SetParams(ctx, authTypes.DefaultParams())
	hApp.CheckpointKeeper.SetParams(ctx, checkpointTypes.DefaultParams())
	hApp.ZenaKeeper.SetParams(ctx, zenaTypes.DefaultParams())

	coins := sdk.Coins{sdk.Coin{Denom: authTypes.FeeToken, Amount: defaultBalance}}
	acc := authTypes.NewBaseAccount(irisAddress,
		coins,
		pubKey,
		0,
		0)

	hApp.AccountKeeper.SetAccount(ctx, acc)
	return hApp, ctx, cosmosCtx.NewCLIContext().WithCodec(hApp.Codec())
}

func prepareMockData(t *testing.T) *gomock.Controller {
	t.Helper()

	mockCtrl := gomock.NewController(t)

	mockHttpClient := helperMocks.NewMockHTTPClient(mockCtrl)
	res := prepareResponse(getAccountResponse)
	defer res.Body.Close()
	mockHttpClient.EXPECT().Get(getAccountUrl).Return(res, nil).AnyTimes()
	helper.Client = mockHttpClient
	return mockCtrl
}

func updateMockData(t *testing.T) *gomock.Controller {
	t.Helper()

	mockCtrl := gomock.NewController(t)

	mockHttpClient := helperMocks.NewMockHTTPClient(mockCtrl)
	res := prepareResponse(getAccountUpdatedResponse)
	defer res.Body.Close()
	mockHttpClient.EXPECT().Get(getAccountUrl).Return(res, nil).AnyTimes()
	helper.Client = mockHttpClient
	return mockCtrl
}

func prepareResponse(body string) *http.Response {
	return &http.Response{
		Status:           "200 OK",
		StatusCode:       200,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             newResettableReadCloser(body),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}
}

// resettableReadCloser resets the reader to the beginning of the data when Close is called.
// this is useful for reusing the response body more than once in tests.
type resettableReadCloser struct {
	data []byte
	r    io.Reader
}

func newResettableReadCloser(body string) *resettableReadCloser {
	return &resettableReadCloser{
		data: []byte(body),
		r:    bytes.NewReader([]byte(body)),
	}
}

func (r *resettableReadCloser) Read(p []byte) (n int, err error) {
	return r.r.Read(p)
}

func (r *resettableReadCloser) Close() error {
	r.r = bytes.NewReader(r.data)
	return nil
}
