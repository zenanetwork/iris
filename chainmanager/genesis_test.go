package chainmanager_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/chainmanager"
	"github.com/zenanetwork/iris/chainmanager/types"
)

// GenesisTestSuite integrate test suite context object
type GenesisTestSuite struct {
	suite.Suite

	app *app.IrisApp
	ctx sdk.Context
}

// SetupTest setup necessary things for genesis test
func (suite *GenesisTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(true)
}

// TestGenesisTestSuite
func TestGenesisTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(GenesisTestSuite))
}

// TestInitExportGenesis test import and export genesis state
func (suite *GenesisTestSuite) TestInitExportGenesis() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	params := types.DefaultParams()

	genesisState := types.GenesisState{
		Params: params,
	}
	chainmanager.InitGenesis(ctx, app.ChainKeeper, genesisState)

	actualParams := chainmanager.ExportGenesis(ctx, app.ChainKeeper)
	require.Equal(t, genesisState, actualParams)
}
