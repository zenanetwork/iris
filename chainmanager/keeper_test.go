package chainmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/chainmanager/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.IrisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)
}

func TestKeeperTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KeeperTestSuite))
}

// Tests

func (suite *KeeperTestSuite) TestParamsGetterSetter() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	params := types.DefaultParams()

	app.ChainKeeper.SetParams(ctx, params)

	actualParams := app.ChainKeeper.GetParams(ctx)

	require.Equal(t, params, actualParams)
}
