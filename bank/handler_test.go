package bank_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zenanetwork/iris/app"
	authTypes "github.com/zenanetwork/iris/auth/types"
	"github.com/zenanetwork/iris/bank"
	"github.com/zenanetwork/iris/bank/types"
	"github.com/zenanetwork/iris/helper/mocks"
	hmTypes "github.com/zenanetwork/iris/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app            *app.IrisApp
	ctx            sdk.Context
	handler        sdk.Handler
	contractCaller mocks.IContractCaller
}

// SetupTest setup all necessary things for querier testing
func (suite *HandlerTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)

	suite.contractCaller = mocks.IContractCaller{}
	suite.handler = bank.NewHandler(suite.app.BankKeeper, &suite.contractCaller)
}

// TestHandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestHandleMsgUnknown() {
	t, _, ctx := suite.T(), suite.app, suite.ctx

	result := suite.handler(ctx, nil)
	require.False(t, result.IsOK())
}

func (suite *HandlerTestSuite) TestHandlerMsgSend() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	amount := int64(10000000)
	from := hmTypes.HexToIrisAddress("123")
	to := hmTypes.HexToIrisAddress("456")
	_, err := app.BankKeeper.AddCoins(ctx, from, sdk.NewCoins(sdk.NewCoin(authTypes.FeeToken, sdk.NewInt(amount*10))))
	require.NoError(t, err)

	msgSend := types.NewMsgSend(
		from,
		to,
		sdk.NewCoins(sdk.NewCoin(authTypes.FeeToken, sdk.NewInt(amount))),
	)
	result := suite.handler(ctx, msgSend)
	require.True(t, result.IsOK(), "Expected New msg to be sent")

	fromAcc := app.BankKeeper.GetCoins(ctx, to)
	require.Less(t, fromAcc.AmountOf(authTypes.FeeToken).Int64(), sdk.NewInt(amount*10).Int64())

	toAcc := app.BankKeeper.GetCoins(ctx, to)
	require.Equal(t, sdk.NewInt(amount), toAcc.AmountOf(authTypes.FeeToken))
}
