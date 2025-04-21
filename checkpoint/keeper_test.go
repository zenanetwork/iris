package checkpoint_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/checkpoint"
	hmTypes "github.com/zenanetwork/iris/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.IrisApp
	ctx sdk.Context
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app, suite.ctx, _ = createTestApp(false)
}

func TestKeeperTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestAddCheckpoint() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	headerBlockNumber := uint64(2000)
	startBlock := uint64(0)
	endBlock := uint64(256)
	rootHash := hmTypes.HexToIrisHash("123")
	proposerAddress := hmTypes.HexToIrisAddress("123")
	timestamp := uint64(time.Now().Unix())
	zenaChainId := "1234"

	Checkpoint := hmTypes.CreateBlock(
		startBlock,
		endBlock,
		rootHash,
		proposerAddress,
		zenaChainId,
		timestamp,
	)
	err := keeper.AddCheckpoint(ctx, headerBlockNumber, Checkpoint)
	require.NoError(t, err)

	result, err := keeper.GetCheckpointByNumber(ctx, headerBlockNumber)
	require.NoError(t, err)
	require.Equal(t, startBlock, result.StartBlock)
	require.Equal(t, endBlock, result.EndBlock)
	require.Equal(t, rootHash, result.RootHash)
	require.Equal(t, zenaChainId, result.ZenaChainID)
	require.Equal(t, proposerAddress, result.Proposer)
	require.Equal(t, timestamp, result.TimeStamp)
}

func (suite *KeeperTestSuite) TestGetCheckpointList() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	count := 5

	startBlock := uint64(0)
	endBlock := uint64(0)

	for i := 0; i < count; i++ {
		headerBlockNumber := uint64(i) + 1

		startBlock = startBlock + endBlock
		endBlock = endBlock + uint64(255)
		rootHash := hmTypes.HexToIrisHash("123")
		proposerAddress := hmTypes.HexToIrisAddress("123")
		timestamp := uint64(time.Now().Unix()) + uint64(i)
		zenaChainId := "1234"

		Checkpoint := hmTypes.CreateBlock(
			startBlock,
			endBlock,
			rootHash,
			proposerAddress,
			zenaChainId,
			timestamp,
		)

		err := keeper.AddCheckpoint(ctx, headerBlockNumber, Checkpoint)
		require.NoError(t, err)

		keeper.UpdateACKCount(ctx)
	}

	result, err := keeper.GetCheckpointList(ctx, uint64(1), uint64(20))
	require.NoError(t, err)
	require.LessOrEqual(t, count, len(result))

	for i := range count {
		require.Equal(t, uint64(i+1), result[i].ID)
	}
}

func (suite *KeeperTestSuite) TestHasStoreValue() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	key := checkpoint.ACKCountKey
	result := keeper.HasStoreValue(ctx, key)
	require.True(t, result)
}

func (suite *KeeperTestSuite) TestFlushCheckpointBuffer() {
	t, app, ctx := suite.T(), suite.app, suite.ctx

	keeper := app.CheckpointKeeper
	key := checkpoint.BufferCheckpointKey

	keeper.FlushCheckpointBuffer(ctx)

	result := keeper.HasStoreValue(ctx, key)
	require.False(t, result)
}
