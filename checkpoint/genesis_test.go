package checkpoint_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/checkpoint"
	"github.com/zenanetwork/iris/checkpoint/types"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/types/simulation"
)

type GenesisTestSuite struct {
	suite.Suite

	app *app.IrisApp
	ctx sdk.Context
}

// SetupTest setup necessary things for genesis test
func (suite *GenesisTestSuite) SetupTest() {
	suite.app, suite.ctx, _ = createTestApp(true)
}

// TestGenesisTestSuite
func TestGenesisTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(GenesisTestSuite))
}

func (suite *GenesisTestSuite) TestInitExportGenesis() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	lastNoACK := simulation.RandIntBetween(r1, 1, 5)
	ackCount := simulation.RandIntBetween(r1, 1, 5)
	startBlock := uint64(0)
	endBlock := uint64(256)
	rootHash := hmTypes.HexToIrisHash("123")

	proposerAddress := hmTypes.HexToIrisAddress("123")
	timestamp := uint64(time.Now().Unix())
	zenaChainId := "1234"

	bufferedCheckpoint := hmTypes.CreateBlock(
		startBlock,
		endBlock,
		rootHash,
		proposerAddress,
		zenaChainId,
		timestamp,
	)

	checkpoints := make([]hmTypes.Checkpoint, ackCount)

	for i := range checkpoints {
		checkpoints[i] = bufferedCheckpoint
	}

	params := types.DefaultParams()
	genesisState := types.NewGenesisState(
		params,
		&bufferedCheckpoint,
		uint64(lastNoACK),
		uint64(ackCount),
		checkpoints,
	)

	checkpoint.InitGenesis(ctx, app.CheckpointKeeper, genesisState)

	actualParams := checkpoint.ExportGenesis(ctx, app.CheckpointKeeper)

	require.Equal(t, genesisState.AckCount, actualParams.AckCount)
	require.Equal(t, genesisState.BufferedCheckpoint, actualParams.BufferedCheckpoint)
	require.Equal(t, genesisState.LastNoACK, actualParams.LastNoACK)
	require.Equal(t, genesisState.Params, actualParams.Params)
	require.LessOrEqual(t, len(actualParams.Checkpoints), len(genesisState.Checkpoints))
}
