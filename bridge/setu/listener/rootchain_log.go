package listener

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/go-zenanet/accounts/abi"
	"github.com/zenanetwork/go-zenanet/core/types"

	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/contracts/stakinginfo"
	"github.com/zenanetwork/iris/contracts/statesender"
	"github.com/zenanetwork/iris/helper"
)

// handleLog handles the given log
func (rl *RootChainListener) handleLog(vLog types.Log, selectedEvent *abi.Event) {
	rl.Logger.Debug("ReceivedEvent", "eventname", selectedEvent.Name)

	switch selectedEvent.Name {
	case "NewHeaderBlock":
		rl.handleNewHeaderBlockLog(vLog, selectedEvent)
	case "Staked":
		rl.handleStakedLog(vLog, selectedEvent)
	case "StakeUpdate":
		rl.handleStakeUpdateLog(vLog, selectedEvent)
	case "SignerChange":
		rl.handleSignerChangeLog(vLog, selectedEvent)
	case "UnstakeInit":
		rl.handleUnstakeInitLog(vLog, selectedEvent)
	case "StateSynced":
		rl.handleStateSyncedLog(vLog, selectedEvent)
	case "TopUpFee":
		rl.handleTopUpFeeLog(vLog, selectedEvent)
	case "Slashed":
		rl.handleSlashedLog(vLog, selectedEvent)
	case "UnJailed":
		rl.handleUnJailedLog(vLog, selectedEvent)
	}
}

func (rl *RootChainListener) handleNewHeaderBlockLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, selectedEvent); isCurrentValidator {
		rl.SendTaskWithDelay("sendCheckpointAckToIris", selectedEvent.Name, logBytes, delay, selectedEvent)
	}
}

func (rl *RootChainListener) handleStakedLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	pubkey := helper.GetPubKey()

	event := new(stakinginfo.StakinginfoStaked)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if !util.IsPubKeyFirstByteValid(pubkey[0:1]) {
		rl.Logger.Error("public key first byte mismatch", "expected", "0x04", "received", pubkey[0:1])
	}

	if bytes.Equal(event.SignerPubkey, pubkey[1:]) {
		// topup has to be processed first before validator join. so adding delay.
		delay := util.TaskDelayBetweenEachVal
		rl.SendTaskWithDelay("sendValidatorJoinToIris", selectedEvent.Name, logBytes, delay, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		// topup has to be processed first before validator join. so adding delay.
		delay = delay + util.TaskDelayBetweenEachVal
		rl.SendTaskWithDelay("sendValidatorJoinToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleStakeUpdateLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	event := new(stakinginfo.StakinginfoStakeUpdate)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if util.IsEventSender(rl.cliCtx, event.ValidatorId.Uint64()) {
		rl.SendTaskWithDelay("sendStakeUpdateToIris", selectedEvent.Name, logBytes, 0, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendStakeUpdateToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleSignerChangeLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	pubkey := helper.GetPubKey()

	event := new(stakinginfo.StakinginfoSignerChange)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if bytes.Equal(event.SignerPubkey, pubkey[1:]) && util.IsPubKeyFirstByteValid(pubkey[0:1]) {
		rl.SendTaskWithDelay("sendSignerChangeToIris", selectedEvent.Name, logBytes, 0, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendSignerChangeToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleUnstakeInitLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	event := new(stakinginfo.StakinginfoUnstakeInit)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if util.IsEventSender(rl.cliCtx, event.ValidatorId.Uint64()) {
		rl.SendTaskWithDelay("sendUnstakeInitToIris", selectedEvent.Name, logBytes, 0, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendUnstakeInitToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleStateSyncedLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	event := new(statesender.StatesenderStateSynced)
	if err = helper.UnpackLog(rl.stateSenderAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	rl.Logger.Info("StateSyncedEvent: detected", "stateSyncId", event.Id)

	if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendStateSyncedToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleTopUpFeeLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	event := new(stakinginfo.StakinginfoTopUpFee)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if bytes.Equal(event.User.Bytes(), helper.GetAddress()) {
		rl.SendTaskWithDelay("sendTopUpFeeToIris", selectedEvent.Name, logBytes, 0, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendTopUpFeeToIris", selectedEvent.Name, logBytes, delay, event)
	}
}

func (rl *RootChainListener) handleSlashedLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, selectedEvent); isCurrentValidator {
		rl.SendTaskWithDelay("sendTickAckToIris", selectedEvent.Name, logBytes, delay, selectedEvent)
	}
}

func (rl *RootChainListener) handleUnJailedLog(vLog types.Log, selectedEvent *abi.Event) {
	logBytes, err := jsoniter.ConfigFastest.Marshal(vLog)
	if err != nil {
		rl.Logger.Error("Failed to marshal log", "Error", err)
	}

	event := new(stakinginfo.StakinginfoUnJailed)
	if err = helper.UnpackLog(rl.stakingInfoAbi, event, selectedEvent.Name, &vLog); err != nil {
		rl.Logger.Error("Error while parsing event", "name", selectedEvent.Name, "error", err)
	}

	if util.IsEventSender(rl.cliCtx, event.ValidatorId.Uint64()) {
		rl.SendTaskWithDelay("sendUnjailToIris", selectedEvent.Name, logBytes, 0, event)
	} else if isCurrentValidator, delay := util.CalculateTaskDelay(rl.cliCtx, event); isCurrentValidator {
		rl.SendTaskWithDelay("sendUnjailToIris", selectedEvent.Name, logBytes, delay, event)
	}
}
