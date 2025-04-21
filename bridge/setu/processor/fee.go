package processor

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/go-zenanet/accounts/abi"
	"github.com/zenanetwork/go-zenanet/core/types"

	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/contracts/stakinginfo"
	"github.com/zenanetwork/iris/helper"
	topupTypes "github.com/zenanetwork/iris/topup/types"
	hmTypes "github.com/zenanetwork/iris/types"
)

// FeeProcessor - process fee related events
type FeeProcessor struct {
	BaseProcessor
	stakingInfoAbi *abi.ABI
}

// NewFeeProcessor - add  abi to clerk processor
func NewFeeProcessor(stakingInfoAbi *abi.ABI) *FeeProcessor {
	return &FeeProcessor{
		stakingInfoAbi: stakingInfoAbi,
	}
}

// Start starts new block subscription
func (fp *FeeProcessor) Start() error {
	fp.Logger.Info("Starting")
	return nil
}

// RegisterTasks - Registers clerk related tasks with machinery
func (fp *FeeProcessor) RegisterTasks() {
	fp.Logger.Info("Registering fee related tasks")

	if err := fp.queueConnector.Server.RegisterTask("sendTopUpFeeToIris", fp.sendTopUpFeeToIris); err != nil {
		fp.Logger.Error("RegisterTasks | sendTopUpFeeToIris", "error", err)
	}
}

// processTopupFeeEvent - processes topup fee event
func (fp *FeeProcessor) sendTopUpFeeToIris(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		fp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoTopUpFee)
	if err := helper.UnpackLog(fp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		fp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		if isOld, _ := fp.isOldTx(fp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.TopupEvent, event); isOld {
			fp.Logger.Info("Ignoring task to send topup to iris as already processed",
				"event", eventName,
				"user", event.User,
				"Fee", event.Fee,
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		fp.Logger.Info("✅ sending topup to iris",
			"event", eventName,
			"user", event.User,
			"Fee", event.Fee,
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// create msg checkpoint ack message
		msg := topupTypes.NewMsgTopup(helper.GetFromAddress(fp.cliCtx), hmTypes.BytesToIrisAddress(event.User.Bytes()), sdk.NewIntFromBigInt(event.Fee), hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()), uint64(vLog.Index), vLog.BlockNumber)

		// return broadcast to iris
		txRes, err := fp.txBroadcaster.BroadcastToIris(msg, event)
		if err != nil {
			fp.Logger.Error("Error while broadcasting TopupFee msg to iris", "msg", msg, "error", err)
			return err
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			fp.Logger.Error("topup tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return fmt.Errorf("topup tx failed, tx response code: %v", txRes.Code)
		}
	}

	return nil
}
