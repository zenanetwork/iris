package checkpoint_test

import (
	"math/big"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	zenaCommon "github.com/zenanetwork/go-zenanet/common"

	"github.com/zenanetwork/iris/app"
	cmTypes "github.com/zenanetwork/iris/chainmanager/types"
	"github.com/zenanetwork/iris/checkpoint"
	chSim "github.com/zenanetwork/iris/checkpoint/simulation"
	"github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/common"
	errs "github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/contracts/rootchain"
	"github.com/zenanetwork/iris/helper/mocks"
	hmTypes "github.com/zenanetwork/iris/types"
)

// SideHandlerTestSuite integrate test suite context object
type SideHandlerTestSuite struct {
	suite.Suite

	app            *app.IrisApp
	ctx            sdk.Context
	sideHandler    hmTypes.SideTxHandler
	postHandler    hmTypes.PostTxHandler
	contractCaller mocks.IContractCaller
	r              *rand.Rand
}

func (suite *SideHandlerTestSuite) SetupTest() {
	suite.app, suite.ctx, _ = createTestApp(false)
	suite.contractCaller = mocks.IContractCaller{}
	suite.sideHandler = checkpoint.NewSideTxHandler(suite.app.CheckpointKeeper, &suite.contractCaller)
	suite.postHandler = checkpoint.NewPostTxHandler(suite.app.CheckpointKeeper, &suite.contractCaller)

	// random generator
	s1 := rand.NewSource(time.Now().UnixNano())
	suite.r = rand.New(s1)
}

func TestSideHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SideHandlerTestSuite))
}

//
// Test cases
//

func (suite *SideHandlerTestSuite) TestSideHandler() {
	t, ctx := suite.T(), suite.ctx

	// side handler
	result := suite.sideHandler(ctx, nil)
	require.Equal(t, uint32(sdk.CodeUnknownRequest), result.Code)
	require.Equal(t, abci.SideTxResultType_Skip, result.Result)
}

// test handler for message
func (suite *HandlerTestSuite) TestHandleMsgCheckpointAdjustSuccess() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	checkpoint := hmTypes.Checkpoint{
		Proposer:    hmTypes.HexToIrisAddress("123"),
		StartBlock:  0,
		EndBlock:    256,
		RootHash:    hmTypes.HexToIrisHash("123"),
		ZenaChainID: "testchainid",
		TimeStamp:   1,
	}
	err := keeper.AddCheckpoint(ctx, 1, checkpoint)
	require.NoError(t, err)

	checkpointAdjust := types.MsgCheckpointAdjust{
		HeaderIndex: 1,
		Proposer:    hmTypes.HexToIrisAddress("456"),
		StartBlock:  0,
		EndBlock:    512,
		RootHash:    hmTypes.HexToIrisHash("456"),
	}
	rootchainInstance := &rootchain.Rootchain{}
	suite.contractCaller.On("GetRootChainInstance", mock.Anything).Return(rootchainInstance, nil)
	suite.contractCaller.On("GetHeaderInfo", mock.Anything, mock.Anything, mock.Anything).Return(zenaCommon.HexToHash("456"), uint64(0), uint64(512), uint64(1), hmTypes.HexToIrisAddress("456"), nil)

	suite.handler(ctx, checkpointAdjust)
	sideResult := suite.sideHandler(ctx, checkpointAdjust)
	suite.postHandler(ctx, checkpointAdjust, sideResult.Result)

	responseCheckpoint, _ := keeper.GetCheckpointByNumber(ctx, 1)
	require.Equal(t, responseCheckpoint.EndBlock, uint64(512))
	require.Equal(t, responseCheckpoint.Proposer, hmTypes.HexToIrisAddress("456"))
	require.Equal(t, responseCheckpoint.RootHash, hmTypes.HexToIrisHash("456"))
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAdjustSameCheckpointAsRootChain() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	checkpoint := hmTypes.Checkpoint{
		Proposer:    hmTypes.HexToIrisAddress("123"),
		StartBlock:  0,
		EndBlock:    256,
		RootHash:    hmTypes.HexToIrisHash("123"),
		ZenaChainID: "testchainid",
		TimeStamp:   1,
	}
	err := keeper.AddCheckpoint(ctx, 1, checkpoint)
	require.NoError(t, err)

	checkpointAdjust := types.MsgCheckpointAdjust{
		HeaderIndex: 1,
		Proposer:    hmTypes.HexToIrisAddress("123"),
		StartBlock:  0,
		EndBlock:    256,
		RootHash:    hmTypes.HexToIrisHash("456"),
	}
	rootchainInstance := &rootchain.Rootchain{}
	suite.contractCaller.On("GetRootChainInstance", mock.Anything).Return(rootchainInstance, nil)
	suite.contractCaller.On("GetHeaderInfo", mock.Anything, mock.Anything, mock.Anything).Return(zenaCommon.HexToHash("123"), uint64(0), uint64(256), uint64(1), hmTypes.HexToIrisAddress("123"), nil)

	suite.handler(ctx, checkpointAdjust)
	sideResult := suite.sideHandler(ctx, checkpointAdjust)
	require.Equal(t, sideResult.Code, uint32(common.CodeCheckpointAlreadyExists))
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAdjustNotSameCheckpointAsRootChain() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	checkpoint := hmTypes.Checkpoint{
		Proposer:    hmTypes.HexToIrisAddress("123"),
		StartBlock:  0,
		EndBlock:    256,
		RootHash:    hmTypes.HexToIrisHash("123"),
		ZenaChainID: "testchainid",
		TimeStamp:   1,
	}
	err := keeper.AddCheckpoint(ctx, 1, checkpoint)
	require.NoError(t, err)

	checkpointAdjust := types.MsgCheckpointAdjust{
		HeaderIndex: 1,
		Proposer:    hmTypes.HexToIrisAddress("123"),
		StartBlock:  0,
		EndBlock:    256,
		RootHash:    hmTypes.HexToIrisHash("123"),
	}

	rootchainInstance := &rootchain.Rootchain{}
	suite.contractCaller.On("GetRootChainInstance", mock.Anything).Return(rootchainInstance, nil)
	suite.contractCaller.On("GetHeaderInfo", mock.Anything, mock.Anything, mock.Anything).Return(zenaCommon.HexToHash("222"), uint64(0), uint64(256), uint64(1), hmTypes.HexToIrisAddress("123"), nil)

	result := suite.sideHandler(ctx, checkpointAdjust)
	require.Equal(t, result.Code, uint32(common.CodeCheckpointAlreadyExists))
}

func (suite *SideHandlerTestSuite) TestSideHandleMsgCheckpoint() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)

	header, err := chSim.GenRandCheckpoint(start, maxSize, params.MaxCheckpointLength)
	require.NoError(t, err)

	zenaChainId := "1234"

	suite.Run("Success", func() {
		suite.contractCaller = mocks.IContractCaller{}

		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		suite.contractCaller.On("CheckIfBlocksExist", header.EndBlock+cmTypes.DefaultMaticchainTxConfirmations).Return(true)
		suite.contractCaller.On("GetRootHash", header.StartBlock, header.EndBlock, uint64(1024)).Return(header.RootHash.Bytes(), nil)

		result := suite.sideHandler(ctx, msgCheckpoint)
		require.Equal(t, uint32(sdk.CodeOK), result.Code, "Side tx handler should be success")

		bufferedHeader, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, bufferedHeader, "Should not store state")
	})

	suite.Run("No Roothash", func() {
		suite.contractCaller = mocks.IContractCaller{}

		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		suite.contractCaller.On("CheckIfBlocksExist", header.EndBlock+cmTypes.DefaultMaticchainTxConfirmations).Return(true)
		suite.contractCaller.On("GetRootHash", header.StartBlock, header.EndBlock, uint64(1024)).Return(nil, nil)

		result := suite.sideHandler(ctx, msgCheckpoint)
		require.NotEqual(t, uint32(sdk.CodeOK), result.Code, "Side tx handler should Fail")
		require.Equal(t, abci.SideTxResultType_Skip, result.Result, "Result should be `skip`")
		require.Equal(t, uint32(common.CodeInvalidBlockInput), result.Code)

		bufferedHeader, err := keeper.GetCheckpointFromBuffer(ctx)
		require.Error(t, err)
		require.Nil(t, bufferedHeader, "Should not store state")
	})

	suite.Run("invalid checkpoint", func() {
		suite.contractCaller = mocks.IContractCaller{}

		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		suite.contractCaller.On("CheckIfBlocksExist", header.EndBlock+cmTypes.DefaultMaticchainTxConfirmations).Return(true)
		suite.contractCaller.On("GetRootHash", header.StartBlock, header.EndBlock, uint64(1024)).Return([]byte{1}, nil)

		result := suite.sideHandler(ctx, msgCheckpoint)
		require.NotEqual(t, uint32(sdk.CodeOK), result.Code, "Side tx handler should fail")
		require.Equal(t, uint32(common.CodeInvalidBlockInput), result.Code)
	})
}

func (suite *SideHandlerTestSuite) TestSideHandleMsgCheckpointAck() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)

	header, _ := chSim.GenRandCheckpoint(start, maxSize, params.MaxCheckpointLength)
	headerId := uint64(1)

	suite.Run("Success", func() {
		suite.contractCaller = mocks.IContractCaller{}

		// prepare ack msg
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			uint64(1),
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)
		rootchainInstance := &rootchain.Rootchain{}

		suite.contractCaller.On("GetRootChainInstance", mock.Anything).Return(rootchainInstance, nil)
		suite.contractCaller.On("GetHeaderInfo", headerId, rootchainInstance, params.ChildBlockInterval).Return(header.RootHash.EthHash(), header.StartBlock, header.EndBlock, header.TimeStamp, header.Proposer, nil)

		result := suite.sideHandler(ctx, msgCheckpointAck)
		require.Equal(t, uint32(sdk.CodeOK), result.Code, "Side tx handler should be success")
		require.Equal(t, abci.SideTxResultType_Yes, result.Result, "Result should be `yes`")
	})

	suite.Run("No HeaderInfo", func() {
		suite.contractCaller = mocks.IContractCaller{}

		// prepare ack msg
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			uint64(1),
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			hmTypes.HexToIrisHash("123"),
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)
		rootchainInstance := &rootchain.Rootchain{}

		suite.contractCaller.On("GetRootChainInstance", mock.Anything).Return(rootchainInstance, nil)
		suite.contractCaller.On("GetHeaderInfo", headerId, rootchainInstance, params.ChildBlockInterval).Return(nil, header.StartBlock, header.EndBlock, header.TimeStamp, header.Proposer, nil)

		result := suite.sideHandler(ctx, msgCheckpointAck)
		require.NotEqual(t, uint32(sdk.CodeOK), result.Code, "Side tx handler should fail")
		require.Equal(t, abci.SideTxResultType_Skip, result.Result, "Result should skip")
	})
}

func (suite *SideHandlerTestSuite) TestPostHandler() {
	t, ctx := suite.T(), suite.ctx

	// side handler
	result := suite.postHandler(ctx, nil, abci.SideTxResultType_Yes)
	require.False(t, result.IsOK(), "Post handler should fail")
	require.Equal(t, sdk.CodeUnknownRequest, result.Code)
}

func (suite *SideHandlerTestSuite) TestPostHandleMsgCheckpoint() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	stakingKeeper := app.StakingKeeper
	topupKeeper := app.TopupKeeper
	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)
	dividendAccount := hmTypes.DividendAccount{
		User:      hmTypes.HexToIrisAddress("123"),
		FeeAmount: big.NewInt(0).String(),
	}
	err := topupKeeper.AddDividendAccount(ctx, dividendAccount)
	require.NoError(t, err)

	// check valid checkpoint
	// generate proposer for validator set
	chSim.LoadValidatorSet(t, 2, stakingKeeper, ctx, false, 10, 0)
	stakingKeeper.IncrementAccum(ctx, 1)

	lastCheckpoint, err := keeper.GetLastCheckpoint(ctx)
	if err == nil {
		start = start + lastCheckpoint.EndBlock + 1
	}

	header, err := chSim.GenRandCheckpoint(start, maxSize, params.MaxCheckpointLength)
	require.NoError(t, err)

	// add current proposer to header
	header.Proposer = stakingKeeper.GetValidatorSet(ctx).Proposer.Signer

	zenaChainId := "1234"

	suite.Run("Failure", func() {
		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_No)
		require.True(t, !result.IsOK(), errs.CodeToDefaultMsg(result.Code))

		bufferedHeader, err := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, bufferedHeader)
		require.Error(t, err)
	})

	suite.Run("Success", func() {
		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		bufferedHeader, err := keeper.GetCheckpointFromBuffer(ctx)
		require.Equal(t, bufferedHeader.StartBlock, header.StartBlock)
		require.Equal(t, bufferedHeader.EndBlock, header.EndBlock)
		require.Equal(t, bufferedHeader.RootHash, header.RootHash)
		require.Equal(t, bufferedHeader.Proposer, header.Proposer)
		require.Equal(t, bufferedHeader.ZenaChainID, header.ZenaChainID)
		require.Empty(t, err, "Unable to set checkpoint from buffer, Error: %v", err)
	})

	suite.Run("Replay", func() {
		// create checkpoint msg
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			zenaChainId,
		)

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.False(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)
		require.Equal(t, common.CodeNoACK, result.Code)
	})
}

func (suite *SideHandlerTestSuite) TestPostHandleMsgCheckpointAck() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)
	header, _ := chSim.GenRandCheckpoint(start, maxSize, params.MaxCheckpointLength)
	// generate proposer for validator set
	chSim.LoadValidatorSet(t, 2, app.StakingKeeper, ctx, false, 10, 0)
	app.StakingKeeper.IncrementAccum(ctx, 1)

	// send ack
	checkpointNumber := uint64(1)

	suite.Run("Failure", func() {
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result := suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_No)
		require.True(t, !result.IsOK(), errs.CodeToDefaultMsg(result.Code))

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)
	})

	suite.Run("Success", func() {
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			header.RootHash,
			"1234",
		)

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)
	})

	suite.Run("Replay", func() {
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result := suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.False(t, result.IsOK())
		require.Equal(t, common.CodeInvalidACK, result.Code)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)
	})

	suite.Run("InvalidEndBlock", func() {
		suite.contractCaller = mocks.IContractCaller{}
		header2, _ := chSim.GenRandCheckpoint(header.EndBlock+1, maxSize, params.MaxCheckpointLength)
		checkpointNumber = checkpointNumber + 1
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header2.Proposer,
			header2.StartBlock,
			header2.EndBlock,
			header2.RootHash,
			header2.RootHash,
			"1234",
		)

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header2.Proposer,
			header2.StartBlock,
			header2.EndBlock,
			header2.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)
	})

	suite.Run("Before Aalborg fork-BufferCheckpoint more than Ack", func() {
		latestCheckpoint, err := keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		header3, _ := chSim.GenRandCheckpoint(latestCheckpoint.EndBlock+1, maxSize, params.MaxCheckpointLength)
		checkpointNumber = checkpointNumber + 1

		msgCheckpoint := types.NewMsgCheckpointBlock(
			header3.Proposer,
			header3.StartBlock,
			header3.EndBlock,
			header3.RootHash,
			header3.RootHash,
			"1234",
		)

		ctx = ctx.WithBlockHeight(int64(-1))

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header3.Proposer,
			header3.StartBlock,
			header3.EndBlock-1,
			header3.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)

		latestCheckpoint, err = keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		require.Equal(t, header3.EndBlock-1, latestCheckpoint.EndBlock, "expected latest checkpoint based on ack value")
	})

	suite.Run("Before Aalborg fork-BufferedCheckpoint less than Ack", func() {
		latestCheckpoint, err := keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		header4, _ := chSim.GenRandCheckpoint(latestCheckpoint.EndBlock+1, maxSize, params.MaxCheckpointLength)
		checkpointNumber = checkpointNumber + 1

		msgCheckpoint := types.NewMsgCheckpointBlock(
			header4.Proposer,
			header4.StartBlock,
			header4.EndBlock,
			header4.RootHash,
			header4.RootHash,
			"1234",
		)

		ctx = ctx.WithBlockHeight(int64(-1))

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header4.Proposer,
			header4.StartBlock,
			header4.EndBlock+1,
			header4.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)

		latestCheckpoint, err = keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		require.Equal(t, header4.EndBlock, latestCheckpoint.EndBlock, "expected latest checkpoint based on ack value")
	})

	suite.Run("After Aalborg fork-BufferCheckpoint more than Ack", func() {
		latestCheckpoint, err := keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		header5, _ := chSim.GenRandCheckpoint(latestCheckpoint.EndBlock+1, maxSize, params.MaxCheckpointLength)
		checkpointNumber = checkpointNumber + 1

		msgCheckpoint := types.NewMsgCheckpointBlock(
			header5.Proposer,
			header5.StartBlock,
			header5.EndBlock,
			header5.RootHash,
			header5.RootHash,
			"1234",
		)

		ctx = ctx.WithBlockHeight(int64(1))

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header5.Proposer,
			header5.StartBlock,
			header5.EndBlock-1,
			header5.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)

		latestCheckpoint, err = keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		require.Equal(t, header5.EndBlock-1, latestCheckpoint.EndBlock, "expected latest checkpoint based on ack value")
	})

	suite.Run("After Aalborg fork-BufferCheckpoint less than Ack", func() {
		latestCheckpoint, err := keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		header6, _ := chSim.GenRandCheckpoint(latestCheckpoint.EndBlock+1, maxSize, params.MaxCheckpointLength)
		checkpointNumber = checkpointNumber + 1

		msgCheckpoint := types.NewMsgCheckpointBlock(
			header6.Proposer,
			header6.StartBlock,
			header6.EndBlock,
			header6.RootHash,
			header6.RootHash,
			"1234",
		)

		ctx = ctx.WithBlockHeight(int64(1))

		result := suite.postHandler(ctx, msgCheckpoint, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-checkpoint to be ok, got %v", result)

		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			checkpointNumber,
			header6.Proposer,
			header6.StartBlock,
			header6.EndBlock+1,
			header6.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		result = suite.postHandler(ctx, msgCheckpointAck, abci.SideTxResultType_Yes)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)

		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Nil(t, afterAckBufferedCheckpoint)

		latestCheckpoint, err = keeper.GetLastCheckpoint(ctx)
		require.Nil(t, err)

		require.Equal(t, header6.EndBlock+1, latestCheckpoint.EndBlock, "expected latest checkpoint based on ack value")
	})
}
