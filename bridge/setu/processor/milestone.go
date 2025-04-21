// nolint
package processor

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maticnetwork/heimdall/bridge/setu/util"
	chainmanagerTypes "github.com/maticnetwork/heimdall/chainmanager/types"
	milestoneTypes "github.com/maticnetwork/heimdall/checkpoint/types"
	"github.com/maticnetwork/heimdall/helper"
	"github.com/pborman/uuid"

	hmTypes "github.com/maticnetwork/heimdall/types"
)

// milestoneProcessor - process milestone related events
type MilestoneProcessor struct {
	BaseProcessor

	// header listener subscription
	cancelMilestoneService context.CancelFunc
}

// MilestoneContext represents milestone context
type MilestoneContext struct {
	ChainmanagerParams *chainmanagerTypes.Params
}

// Start starts new block subscription
func (mp *MilestoneProcessor) Start() error {
	mp.Logger.Info("Starting")

	// create cancellable context
	milestoneCtx, cancelMilestoneService := context.WithCancel(context.Background())

	mp.cancelMilestoneService = cancelMilestoneService

	// start polling for milestone
	mp.Logger.Info("Start polling for milestone", "milestoneLength", helper.MilestoneLength, "pollInterval", helper.GetConfig().MilestonePollInterval)

	go mp.startPolling(milestoneCtx, helper.MilestoneLength, helper.GetConfig().MilestonePollInterval)
	go mp.startPollingMilestoneTimeout(milestoneCtx, 2*helper.GetConfig().MilestonePollInterval)

	return nil
}

// RegisterTasks - nil
func (mp *MilestoneProcessor) RegisterTasks() {

}

// startPolling - polls iris and checks if new milestone needs to be proposed
func (mp *MilestoneProcessor) startPolling(ctx context.Context, milestoneLength uint64, interval time.Duration) {
	ticker := time.NewTicker(interval)
	// stop ticker when everything done
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := mp.checkAndPropose(milestoneLength)
			if err != nil {
				mp.Logger.Error("Error in proposing the milestone", "error", err)
			}
		case <-ctx.Done():
			mp.Logger.Info("Polling stopped")
			return
		}
	}
}

// sendMilestoneToIris - handles headerblock from maticchain
// 1. check if i am the proposer for next milestone
// 2. check if milestone has to be proposed
// 3. if so, propose milestone to iris.
func (mp *MilestoneProcessor) checkAndPropose(milestoneLength uint64) (err error) {
	//Milestone proposing mechanism will work only after specific block height
	if util.GetBlockHeight(mp.cliCtx) < helper.GetAalborgHardForkHeight() {
		mp.Logger.Debug("Block height Less than fork height", "current block height", util.GetBlockHeight(mp.cliCtx), "milestone hard fork height", helper.GetAalborgHardForkHeight())
		return nil
	}

	// fetch milestone context
	milestoneContext, err := mp.getMilestoneContext()
	if err != nil {
		return err
	}

	//check whether the node is current milestone proposer or not
	isProposer, err := util.IsMilestoneProposer(mp.cliCtx)
	if err != nil {
		mp.Logger.Error("Error checking isProposer in HeaderBlock handler", "error", err)
		return err
	}

	if isProposer {
		result, err := util.GetMilestoneCount(mp.cliCtx)
		if err != nil {
			return err
		}

		if result == nil {
			return errors.New("got nil result while fetching milestone count")
		}

		var start = helper.GetMilestoneBorBlockHeight()

		if result.Count != 0 {
			// fetch latest milestone
			latestMilestone, err := util.GetLatestMilestone(mp.cliCtx)
			if err != nil {
				return err
			}

			if latestMilestone == nil {
				return errors.New("got nil result while fetching latest milestone")
			}

			//start block number should be continuous to the end block of lasted stored milestone
			start = latestMilestone.EndBlock + 1
		}

		//send the milestone to iris chain
		if err := mp.createAndSendMilestoneToIris(milestoneContext, start, milestoneLength); err != nil {
			mp.Logger.Error("Error sending milestone to iris", "error", err)
			return err
		}
	} else {
		mp.Logger.Info("I am not the current milestone proposer")
	}

	return nil
}

// sendMilestoneToIris - creates milestone msg and broadcasts to iris
func (mp *MilestoneProcessor) createAndSendMilestoneToIris(milestoneContext *MilestoneContext, startNum uint64, milestoneLength uint64) error {
	mp.Logger.Debug("Initiating milestone to Iris", "start", startNum, "milestoneLength", milestoneLength)

	blocksConfirmation := helper.MaticChainMilestoneConfirmation

	// Get latest matic block
	block, err := mp.contractConnector.GetMaticChainBlock(nil)
	if err != nil {
		return err
	}

	latestNum := block.Number.Uint64()

	if latestNum < startNum+milestoneLength+blocksConfirmation-1 {
		return fmt.Errorf("Less than milestoneLength  Start=%v Latest Block=%v MilestoneLength=%v MaticChainConfirmation=%v", startNum, latestNum, milestoneLength, blocksConfirmation)
	}

	endNum := latestNum - blocksConfirmation

	//fetch the endBlock+1 number instead of endBlock so that we can directly get the hash of endBlock using parent hash
	block, err = mp.contractConnector.GetMaticChainBlock(big.NewInt(int64(endNum + 1)))
	if err != nil {
		return fmt.Errorf("error while fetching block %d: %w", endNum+1, err)
	}

	endHash := block.ParentHash

	milestoneId := fmt.Sprintf("%s - %s", uuid.NewRandom().String(), hmTypes.BytesToIrisAddress(endHash[:]).String())

	mp.Logger.Info("End block hash", hmTypes.BytesToIrisHash(endHash[:]))

	mp.Logger.Info("✅ Creating and broadcasting new milestone",
		"start", startNum,
		"end", endNum,
		"hash", hmTypes.BytesToIrisHash(endHash[:]),
		"milestoneId", milestoneId,
		"milestoneLength", milestoneLength,
	)

	chainParams := milestoneContext.ChainmanagerParams.ChainParams

	// create and send milestone message
	msg := milestoneTypes.NewMsgMilestoneBlock(
		hmTypes.BytesToIrisAddress(helper.GetAddress()),
		startNum,
		endNum,
		hmTypes.BytesToIrisHash(endHash[:]),
		chainParams.BorChainID,
		milestoneId,
	)

	//broadcast to iris
	txRes, err := mp.txBroadcaster.BroadcastToIris(msg, nil)
	if err != nil {
		mp.Logger.Error("Error while broadcasting milestone to iris", "error", err)
		return err
	}

	if txRes.Code != uint32(sdk.CodeOK) {
		mp.Logger.Error("milestone tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
		return fmt.Errorf("milestone tx failed, tx response code: %v", txRes.Code)

	}
	return nil
}

// startPolling - polls iris and checks if new milestoneTimeout needs to be proposed
func (mp *MilestoneProcessor) startPollingMilestoneTimeout(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	// stop ticker when everything done
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := mp.checkAndProposeMilestoneTimeout()
			if err != nil {
				mp.Logger.Error("Error in proposing the MilestoneTimeout msg", "error", err)
			}
		case <-ctx.Done():
			mp.Logger.Info("Polling stopped")
			ticker.Stop()

			return
		}
	}
}

// sendMilestoneToIris - handles headerblock from maticchain
// 1. check if i am the proposer for next milestone
// 2. check if milestone has to be proposed
// 3. if so, propose milestone to iris.
func (mp *MilestoneProcessor) checkAndProposeMilestoneTimeout() (err error) {
	//Milestone proposing mechanism will work only after specific block height
	if util.GetBlockHeight(mp.cliCtx) < helper.GetAalborgHardForkHeight() {
		mp.Logger.Debug("Block height Less than fork height", "current block height", util.GetBlockHeight(mp.cliCtx), "milestone hard fork height", helper.GetAalborgHardForkHeight())
		return nil
	}

	isMilestoneTimeoutRequired, err := mp.checkIfMilestoneTimeoutIsRequired()
	if err != nil {
		mp.Logger.Debug("Error checking sMilestoneTimeoutRequired while proposing Milestone Timeout ", "error", err)
		return
	}

	if isMilestoneTimeoutRequired {
		var isProposer bool

		//check if the node is the proposer list or not.
		if isProposer, err = util.IsInMilestoneProposerList(mp.cliCtx, 10); err != nil {
			mp.Logger.Error("Error checking IsInMilestoneProposerList while proposing Milestone Timeout ", "error", err)
			return
		}

		// if i am the proposer and NoAck is required, then propose No-Ack
		if isProposer {
			// send Checkpoint No-Ack to iris
			if err = mp.createAndSendMilestoneTimeoutToIris(); err != nil {
				mp.Logger.Error("Error proposing Milestone-Timeout ", "error", err)
				return
			}
		}
	}

	return nil
}

// sendMilestoneTimeoutToIris - creates milestone-timeout msg and broadcasts to iris
func (mp *MilestoneProcessor) createAndSendMilestoneTimeoutToIris() error {
	mp.Logger.Debug("Initiating milestone timeout to Iris")

	mp.Logger.Info("✅ Creating and broadcasting milestone-timeout")

	// create and send milestone message
	msg := milestoneTypes.NewMsgMilestoneTimeout(
		hmTypes.BytesToIrisAddress(helper.GetAddress()),
	)

	// return broadcast to iris
	txRes, err := mp.txBroadcaster.BroadcastToIris(msg, nil)
	if err != nil {
		mp.Logger.Error("Error while broadcasting milestone timeout to iris", "error", err)
		return err
	}

	if txRes.Code != uint32(sdk.CodeOK) {
		mp.Logger.Error("milestone timeout tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
		return fmt.Errorf("milestone timeout tx failed, tx response code: %v", txRes.Code)
	}

	return nil
}

func (mp *MilestoneProcessor) checkIfMilestoneTimeoutIsRequired() (bool, error) {
	latestMilestone, err := util.GetLatestMilestone(mp.cliCtx)
	if err != nil || latestMilestone == nil {
		return false, err
	}

	lastMilestoneEndBlock := latestMilestone.EndBlock
	currentChildBlockNumber, _ := mp.getCurrentChildBlock()

	if err != nil {
		return false, err
	}

	if (currentChildBlockNumber - lastMilestoneEndBlock) > helper.MilestoneBufferLength {
		return true, nil
	}

	return false, nil
}

// getCurrentChildBlock gets the current child block
func (mp *MilestoneProcessor) getCurrentChildBlock() (uint64, error) {
	childBlock, err := mp.contractConnector.GetMaticChainBlock(nil)
	if err != nil {
		return 0, err
	}

	return childBlock.Number.Uint64(), nil
}

func (mp *MilestoneProcessor) getMilestoneContext() (*MilestoneContext, error) {
	chainmanagerParams, err := util.GetChainmanagerParams(mp.cliCtx)
	if err != nil {
		mp.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	return &MilestoneContext{
		ChainmanagerParams: chainmanagerParams,
	}, nil
}

// Stop stops all necessary go routines
func (mp *MilestoneProcessor) Stop() {
	// cancel milestone polling
	mp.cancelMilestoneService()
}
