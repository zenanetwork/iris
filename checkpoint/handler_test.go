package checkpoint_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zenanetwork/iris/app"
	cmTypes "github.com/zenanetwork/iris/chainmanager/types"
	"github.com/zenanetwork/iris/checkpoint"
	chSim "github.com/zenanetwork/iris/checkpoint/simulation"
	"github.com/zenanetwork/iris/checkpoint/types"
	errs "github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/helper/mocks"
	hmTypes "github.com/zenanetwork/iris/types"
)

type HandlerTestSuite struct {
	suite.Suite

	app    *app.IrisApp
	ctx    sdk.Context
	cliCtx context.CLIContext

	handler        sdk.Handler
	sideHandler    hmTypes.SideTxHandler
	postHandler    hmTypes.PostTxHandler
	contractCaller mocks.IContractCaller
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.app, suite.ctx, suite.cliCtx = createTestApp(false)
	suite.contractCaller = mocks.IContractCaller{}
	suite.handler = checkpoint.NewHandler(suite.app.CheckpointKeeper, &suite.contractCaller)
	suite.sideHandler = checkpoint.NewSideTxHandler(suite.app.CheckpointKeeper, &suite.contractCaller)
	suite.postHandler = checkpoint.NewPostTxHandler(suite.app.CheckpointKeeper, &suite.contractCaller)
}

func TestHandlerTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(HandlerTestSuite))
}

func (suite *HandlerTestSuite) TestHandler() {
	t, ctx := suite.T(), suite.ctx

	// side handler
	result := suite.handler(ctx, nil)
	require.False(t, result.IsOK(), "Handler should fail")
}

// test handler for message
func (suite *HandlerTestSuite) TestHandleMsgCheckpoint() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	stakingKeeper := app.StakingKeeper
	topupKeeper := app.TopupKeeper
	start := uint64(0)
	maxSize := uint64(256)
	zenaChainId := "1234"
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

	dividendAccounts := topupKeeper.GetAllDividendAccounts(ctx)
	accRootHash, err := types.GetAccountRootHash(dividendAccounts)
	require.NoError(t, err)

	accountRoot := hmTypes.BytesToIrisHash(accRootHash)

	suite.Run("Success", func() {
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			accountRoot,
			zenaChainId,
		)

		// send checkpoint to handler
		got := suite.handler(ctx, msgCheckpoint)
		require.True(t, got.IsOK(), "expected send-checkpoint to be ok, got %v", got)
		bufferedHeader, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.Empty(t, bufferedHeader, "Should not store state")
	})

	suite.Run("Invalid Proposer", func() {
		header.Proposer = hmTypes.HexToIrisAddress("1234")
		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			accountRoot,
			zenaChainId,
		)

		// send checkpoint to handler
		got := suite.handler(ctx, msgCheckpoint)
		require.True(t, !got.IsOK(), errs.CodeToDefaultMsg(got.Code))
	})

	suite.Run("Checkpoint not in continuity", func() {
		headerId := uint64(10000)

		err = keeper.AddCheckpoint(ctx, headerId, header)
		require.NoError(t, err)

		_, err = keeper.GetCheckpointByNumber(ctx, headerId)
		require.NoError(t, err)

		keeper.UpdateACKCount(ctx)
		lastCheckpoint, err := keeper.GetLastCheckpoint(ctx)
		if err == nil {
			// pass wrong start
			start = start + lastCheckpoint.EndBlock + 2
		}

		msgCheckpoint := types.NewMsgCheckpointBlock(
			header.Proposer,
			start,
			start+256,
			header.RootHash,
			accountRoot,
			zenaChainId,
		)

		// send checkpoint to handler
		got := suite.handler(ctx, msgCheckpoint)
		require.True(t, !got.IsOK(), errs.CodeToDefaultMsg(got.Code))
	})
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAdjustCheckpointBuffer() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	checkpoint := hmTypes.Checkpoint{
		Proposer:   hmTypes.HexToIrisAddress("123"),
		StartBlock: 0,
		EndBlock:   256,
		RootHash:   hmTypes.HexToIrisHash("123"),
		ZenaChainID: "testchainid",
		TimeStamp:  1,
	}

	err := keeper.SetCheckpointBuffer(ctx, checkpoint)
	require.NoError(t, err)

	checkpointAdjust := types.MsgCheckpointAdjust{
		HeaderIndex: 1,
		Proposer:    hmTypes.HexToIrisAddress("456"),
		StartBlock:  0,
		EndBlock:    512,
		RootHash:    hmTypes.HexToIrisHash("456"),
	}

	result := suite.handler(ctx, checkpointAdjust)
	require.False(t, result.IsOK())
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAdjustSameCheckpointAsMsg() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper

	checkpoint := hmTypes.Checkpoint{
		Proposer:   hmTypes.HexToIrisAddress("123"),
		StartBlock: 0,
		EndBlock:   256,
		RootHash:   hmTypes.HexToIrisHash("123"),
		ZenaChainID: "testchainid",
		TimeStamp:  1,
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

	result := suite.handler(ctx, checkpointAdjust)
	require.False(t, result.IsOK())
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAfterBufferTimeOut() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	stakingKeeper := app.StakingKeeper
	topupKeeper := app.TopupKeeper
	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)
	checkpointBufferTime := params.CheckpointBufferTime
	dividendAccount := hmTypes.DividendAccount{
		User:      hmTypes.HexToIrisAddress("123"),
		FeeAmount: big.NewInt(0).String(),
	}
	err := topupKeeper.AddDividendAccount(ctx, dividendAccount)
	require.NoError(t, err)

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

	// send old checkpoint
	res := suite.SendCheckpoint(header)
	require.True(t, res.IsOK(), "expected send-checkpoint to be  ok, got %v", res)

	checkpointBuffer, err := keeper.GetCheckpointFromBuffer(ctx)
	require.NoError(t, err)

	// set time buffered checkpoint timestamp + checkpointBufferTime
	newTime := checkpointBuffer.TimeStamp + uint64(checkpointBufferTime)
	suite.ctx = ctx.WithBlockTime(time.Unix(0, int64(newTime)))

	// send new checkpoint which should replace old one
	got := suite.SendCheckpoint(header)
	require.True(t, got.IsOK(), "expected send-checkpoint to be  ok, got %v", got)
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointExistInBuffer() {
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

	// send old checkpoint
	res := suite.SendCheckpoint(header)

	require.True(t, res.IsOK(), "expected send-checkpoint to be  ok, got %v", res)

	// send checkpoint to handler
	got := suite.SendCheckpoint(header)
	require.True(t, !got.IsOK(), errs.CodeToDefaultMsg(got.Code))
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointAck() {
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

	got := suite.SendCheckpoint(header)
	require.True(t, got.IsOK(), "expected send-checkpoint to be ok, got %v", got)

	bufferedHeader, err := keeper.GetCheckpointFromBuffer(ctx)
	require.NoError(t, err)
	require.NotNil(t, bufferedHeader)

	// send ack
	headerId := uint64(1)

	suite.Run("success", func() {
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			headerId,
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)
		result := suite.handler(ctx, msgCheckpointAck)
		require.True(t, result.IsOK(), "expected send-ack to be ok, got %v", result)
		afterAckBufferedCheckpoint, _ := keeper.GetCheckpointFromBuffer(ctx)
		require.NotNil(t, afterAckBufferedCheckpoint, "should not remove from buffer")
	})

	suite.Run("Invalid start", func() {
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			headerId,
			header.Proposer,
			uint64(123),
			header.EndBlock,
			header.RootHash,
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		got := suite.handler(ctx, msgCheckpointAck)
		require.True(t, !got.IsOK(), errs.CodeToDefaultMsg(got.Code))
	})

	suite.Run("Invalid Roothash", func() {
		msgCheckpointAck := types.NewMsgCheckpointAck(
			hmTypes.HexToIrisAddress("123"),
			headerId,
			header.Proposer,
			header.StartBlock,
			header.EndBlock,
			hmTypes.HexToIrisHash("9887"),
			hmTypes.HexToIrisHash("123123"),
			uint64(1),
		)

		got := suite.handler(ctx, msgCheckpointAck)
		require.True(t, !got.IsOK(), errs.CodeToDefaultMsg(got.Code))
	})
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointNoAck() {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	keeper := app.CheckpointKeeper
	stakingKeeper := app.StakingKeeper
	topupKeeper := app.TopupKeeper
	start := uint64(0)
	maxSize := uint64(256)
	params := keeper.GetParams(ctx)
	checkpointBufferTime := params.CheckpointBufferTime

	dividendAccount := hmTypes.DividendAccount{
		User:      hmTypes.HexToIrisAddress("123"),
		FeeAmount: big.NewInt(0).String(),
	}
	err := topupKeeper.AddDividendAccount(ctx, dividendAccount)
	require.NoError(t, err)

	// check valid checkpoint
	// generate proposer for validator set
	chSim.LoadValidatorSet(t, 4, stakingKeeper, ctx, false, 10, 0)
	stakingKeeper.IncrementAccum(ctx, 1)

	lastCheckpoint, err := keeper.GetLastCheckpoint(ctx)
	if err == nil {
		start = start + lastCheckpoint.EndBlock + 1
	}

	header, err := chSim.GenRandCheckpoint(start, maxSize, params.MaxCheckpointLength)
	require.NoError(t, err)

	// add current proposer to header
	header.Proposer = stakingKeeper.GetValidatorSet(ctx).Proposer.Signer

	got := suite.SendCheckpoint(header)
	require.True(t, got.IsOK(), "expected send-NoAck to be ok, got %v", got)

	// set time lastCheckpoint timestamp + checkpointBufferTime-10
	newTime := lastCheckpoint.TimeStamp + uint64(checkpointBufferTime) - uint64(10)
	suite.ctx = ctx.WithBlockTime(time.Unix(0, int64(newTime)))

	validatorSet := stakingKeeper.GetValidatorSet(ctx)

	//Rotate the list to get the next proposer in line
	validatorSet.IncrementProposerPriority(1)
	noAckProposer := validatorSet.Proposer.Signer

	result := suite.SendNoAck(noAckProposer)
	require.False(t, result.IsOK(), "expected send-NoAck to be false, got %v", true)

	ackCount := keeper.GetACKCount(ctx)
	require.Equal(t, uint64(0), ackCount, "Should not update state")

	// set time lastCheckpoint timestamp + noAckWaitTime
	newTime = lastCheckpoint.TimeStamp + uint64(checkpointBufferTime)
	suite.ctx = ctx.WithBlockTime(time.Unix(0, int64(newTime)))

	//This noAck should false as noAckProposer is invalid, we are passing current
	//checkpoint proposer as noAck proposer
	result = suite.SendNoAck(stakingKeeper.GetValidatorSet(ctx).Proposer.Signer)
	require.False(t, result.IsOK(), "expected send-NoAck to be false , got %v", true)

	//This noAck should return true as noAckProposer is valid
	result = suite.SendNoAck(noAckProposer)
	require.True(t, result.IsOK(), "expected send-NoAck to be true, got %v", false)

	ackCount = keeper.GetACKCount(ctx)
	require.Equal(t, uint64(0), ackCount, "Should not update state")
}

func (suite *HandlerTestSuite) TestHandleMsgCheckpointNoAckBeforeBufferTimeout() {
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

	got := suite.SendCheckpoint(header)
	require.True(t, got.IsOK(), "expected send-checkpoint to be ok, got %v", got)

	validatorSet := stakingKeeper.GetValidatorSet(ctx)

	//Rotate the list to get the next proposer in line
	validatorSet.IncrementProposerPriority(1)
	noAckProposer := validatorSet.Proposer.Signer

	result := suite.SendNoAck(noAckProposer)
	require.True(t, !result.IsOK(), errs.CodeToDefaultMsg(result.Code))
}

func (suite *HandlerTestSuite) SendCheckpoint(header hmTypes.Checkpoint) (res sdk.Result) {
	t, app, ctx := suite.T(), suite.app, suite.ctx
	// keeper := app.CheckpointKeeper
	topupKeeper := app.TopupKeeper

	dividendAccounts := topupKeeper.GetAllDividendAccounts(ctx)
	accRootHash, err := types.GetAccountRootHash(dividendAccounts)
	require.NoError(t, err)

	accountRoot := hmTypes.BytesToIrisHash(accRootHash)

	zenaChainId := "1234"
	// create checkpoint msg
	msgCheckpoint := types.NewMsgCheckpointBlock(
		header.Proposer,
		header.StartBlock,
		header.EndBlock,
		header.RootHash,
		accountRoot,
		zenaChainId,
	)

	suite.contractCaller.On("CheckIfBlocksExist", header.EndBlock+cmTypes.DefaultMaticchainTxConfirmations).Return(true)
	suite.contractCaller.On("GetRootHash", header.StartBlock, header.EndBlock, uint64(1024)).Return(header.RootHash.Bytes(), nil)

	// send checkpoint to handler
	result := suite.handler(ctx, msgCheckpoint)
	sideResult := suite.sideHandler(ctx, msgCheckpoint)
	suite.postHandler(ctx, msgCheckpoint, sideResult.Result)

	return result
}

func (suite *HandlerTestSuite) SendNoAck(noAckProposer hmTypes.IrisAddress) (res sdk.Result) {
	_, _, ctx := suite.T(), suite.app, suite.ctx
	msgNoAck := types.NewMsgCheckpointNoAck(noAckProposer)

	result := suite.handler(ctx, msgNoAck)
	sideResult := suite.sideHandler(ctx, msgNoAck)
	suite.postHandler(ctx, msgNoAck, sideResult.Result)

	return result
}
