package checkpoint

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"

	cmTypes "github.com/zenanetwork/iris/chainmanager/types"
	"github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/common"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
)

// SideHandleMsgMilestone handles MsgMilestone message for external call
func SideHandleMsgMilestone(ctx sdk.Context, k Keeper, msg types.MsgMilestone, contractCaller helper.IContractCaller) (result abci.ResponseDeliverSideTx) {
	// get params
	milestoneLength := helper.MilestoneLength

	// logger
	logger := k.MilestoneLogger(ctx)

	//Get the milestone count
	count := k.GetMilestoneCount(ctx)
	lastMilestone, err := k.GetLastMilestone(ctx)

	if count != uint64(0) && err != nil {
		logger.Error("Error while receiving the last milestone in the side handler")
		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidBlockInput)
	}

	if count != uint64(0) && msg.StartBlock != lastMilestone.EndBlock+1 {
		logger.Error("Milestone is not in continuity to last stored milestone",
			"startBlock", msg.StartBlock,
			"endBlock", msg.EndBlock,
			"hash", msg.Hash,
			"milestoneId", msg.MilestoneID,
			"error", err,
		)

		return common.ErrorSideTx(k.Codespace(), common.CodeInvalidBlockInput)
	}

	// Validating the milestone
	validMilestone, err := types.ValidateMilestone(msg.StartBlock, msg.EndBlock, msg.Hash, msg.MilestoneID, contractCaller, milestoneLength, cmTypes.DefaultMaticchainMilestoneTxConfirmations)
	if err != nil {
		logger.Error("Error validating milestone",
			"startBlock", msg.StartBlock,
			"endBlock", msg.EndBlock,
			"hash", msg.Hash,
			"milestoneId", msg.MilestoneID,
			"error", err,
		)
	} else if validMilestone {
		// vote `yes` if milestone is valid
		result.Result = abci.SideTxResultType_Yes
		return
	}

	logger.Error(
		"Hash is not valid",
		"startBlock", msg.StartBlock,
		"endBlock", msg.EndBlock,
		"hash", msg.Hash,
		"milestoneId", msg.MilestoneID,
	)

	return common.ErrorSideTx(k.Codespace(), common.CodeInvalidBlockInput)
}

// PostHandleMsgMilestone handles msg milestone
func PostHandleMsgMilestone(ctx sdk.Context, k Keeper, msg types.MsgMilestone, sideTxResult abci.SideTxResultType) sdk.Result {
	logger := k.MilestoneLogger(ctx)
	time := ctx.BlockTime().Unix()
	//nolint:gosec
	timeStamp := uint64(time)

	// TX bytes
	txBytes := ctx.TxBytes()
	hash := tmTypes.Tx(txBytes).Hash()

	// Emit event for milestone
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeMilestone,
			sdk.NewAttribute(sdk.AttributeKeyAction, msg.Type()),                              // action
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),            // module name
			sdk.NewAttribute(hmTypes.AttributeKeyTxHash, hmTypes.BytesToIrisHash(hash).Hex()), // tx hash
			sdk.NewAttribute(hmTypes.AttributeKeySideTxResult, sideTxResult.String()),         // result
			sdk.NewAttribute(types.AttributeKeyProposer, msg.Proposer.String()),
			sdk.NewAttribute(types.AttributeKeyStartBlock, strconv.FormatUint(msg.StartBlock, 10)),
			sdk.NewAttribute(types.AttributeKeyEndBlock, strconv.FormatUint(msg.EndBlock, 10)),
			sdk.NewAttribute(types.AttributeKeyHash, msg.Hash.String()),
			sdk.NewAttribute(types.AttributeKeyMilestoneID, msg.MilestoneID),
		),
	})

	// Skip handler if milestone is not approved
	if sideTxResult != abci.SideTxResultType_Yes {
		logger.Debug("Skipping new milestone since side-tx didn't get yes votes", "startBlock", msg.StartBlock, "endBlock", msg.EndBlock, "hash", msg.Hash, "milestoneId", msg.MilestoneID)
		k.SetNoAckMilestone(ctx, msg.MilestoneID)

		return sdk.Result{
			Events: ctx.EventManager().Events(),
		}
	}

	//Get the latest stored milestone from store
	if lastMilestone, err := k.GetLastMilestone(ctx); err == nil { // fetch last milestone from store
		// make sure new milestoen is after tip
		if lastMilestone.EndBlock > msg.StartBlock {
			logger.Error(" Milestone already exists",
				"currentTip", lastMilestone.EndBlock,
				"startBlock", msg.StartBlock,
			)

			k.SetNoAckMilestone(ctx, msg.MilestoneID)

			return sdk.Result{
				Events: ctx.EventManager().Events(),
			}
		}

		// check if new milestone's start block start from current tip
		if lastMilestone.EndBlock+1 != msg.StartBlock {
			logger.Error("milestone not in countinuity",
				"currentTip", lastMilestone.EndBlock,
				"startBlock", msg.StartBlock)

			k.SetNoAckMilestone(ctx, msg.MilestoneID)

			return sdk.Result{
				Events: ctx.EventManager().Events(),
			}
		}
	} else if msg.StartBlock != helper.GetMilestoneZenaBlockHeight() {
		logger.Error("First milestone to start from", "block", helper.GetMilestoneZenaBlockHeight(), "Error", err)

		k.SetNoAckMilestone(ctx, msg.MilestoneID)

		return sdk.Result{
			Events: ctx.EventManager().Events(),
		}

	}

	//Add the milestone to the store
	if err := k.AddMilestone(ctx, hmTypes.Milestone{ // Save milestone to buffer store
		StartBlock:  msg.StartBlock, //Add milestone to store with root hash
		EndBlock:    msg.EndBlock,
		Hash:        msg.Hash,
		Proposer:    msg.Proposer,
		ZenaChainID:  msg.ZenaChainID,
		MilestoneID: msg.MilestoneID,
		TimeStamp:   timeStamp,
	}); err != nil {
		k.SetNoAckMilestone(ctx, msg.MilestoneID)
		logger.Error("Failed to set milestone ", "Error", err)
	}

	return sdk.Result{
		Events: ctx.EventManager().Events(),
	}
}
