package processor

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
	cliContext "github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/go-zenanet/accounts/abi"
	"github.com/zenanetwork/go-zenanet/core/types"

	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/contracts/stakinginfo"
	"github.com/zenanetwork/iris/helper"
	stakingTypes "github.com/zenanetwork/iris/staking/types"
	hmTypes "github.com/zenanetwork/iris/types"
)

const (
	defaultDelayDuration time.Duration = 15 * time.Second
)

// StakingProcessor - process staking related events
type StakingProcessor struct {
	BaseProcessor
	stakingInfoAbi *abi.ABI
}

// NewStakingProcessor - add  abi to staking processor
func NewStakingProcessor(stakingInfoAbi *abi.ABI) *StakingProcessor {
	return &StakingProcessor{
		stakingInfoAbi: stakingInfoAbi,
	}
}

// Start starts new block subscription
func (sp *StakingProcessor) Start() error {
	sp.Logger.Info("Starting")
	return nil
}

// RegisterTasks - Registers staking tasks with machinery
func (sp *StakingProcessor) RegisterTasks() {
	sp.Logger.Info("Registering staking related tasks")

	if err := sp.queueConnector.Server.RegisterTask("sendValidatorJoinToIris", sp.sendValidatorJoinToIris); err != nil {
		sp.Logger.Error("RegisterTasks | sendValidatorJoinToIris", "error", err)
	}

	if err := sp.queueConnector.Server.RegisterTask("sendUnstakeInitToIris", sp.sendUnstakeInitToIris); err != nil {
		sp.Logger.Error("RegisterTasks | sendUnstakeInitToIris", "error", err)
	}

	if err := sp.queueConnector.Server.RegisterTask("sendStakeUpdateToIris", sp.sendStakeUpdateToIris); err != nil {
		sp.Logger.Error("RegisterTasks | sendStakeUpdateToIris", "error", err)
	}

	if err := sp.queueConnector.Server.RegisterTask("sendSignerChangeToIris", sp.sendSignerChangeToIris); err != nil {
		sp.Logger.Error("RegisterTasks | sendSignerChangeToIris", "error", err)
	}
}

func (sp *StakingProcessor) sendValidatorJoinToIris(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoStaked)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		signerPubKey := event.SignerPubkey
		if len(signerPubKey) == 64 {
			signerPubKey = util.AppendPrefix(signerPubKey)
		}
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.StakingEvent, event); isOld {
			sp.Logger.Info("Ignoring task to send validatorjoin to iris as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"activationEpoch", event.ActivationEpoch,
				"nonce", event.Nonce,
				"amount", event.Amount,
				"totalAmount", event.Total,
				"SignerPubkey", hmTypes.NewPubKey(signerPubKey).String(),
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		// if account doesn't exists Retry with delay for topup to process first.
		if _, err := util.GetAccount(sp.cliCtx, hmTypes.IrisAddress(event.Signer)); err != nil {
			sp.Logger.Info(
				"Iris Account doesn't exist. Retrying validator-join after 10 seconds",
				"event", eventName,
				"signer", event.Signer,
			)
			return tasks.NewErrRetryTaskLater("account doesn't exist", util.RetryTaskDelay)
		}

		sp.Logger.Info(
			"✅ Received task to send validatorjoin to iris",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"activationEpoch", event.ActivationEpoch,
			"nonce", event.Nonce,
			"amount", event.Amount,
			"totalAmount", event.Total,
			"SignerPubkey", hmTypes.NewPubKey(signerPubKey).String(),
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator join
		msg := stakingTypes.NewMsgValidatorJoin(
			hmTypes.BytesToIrisAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			event.ActivationEpoch.Uint64(),
			sdk.NewIntFromBigInt(event.Amount),
			hmTypes.NewPubKey(signerPubKey),
			hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to iris
		txRes, err := sp.txBroadcaster.BroadcastToIris(msg, event)
		if err != nil {
			sp.Logger.Error("Error while broadcasting unstakeInit to iris", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			sp.Logger.Error("validator-join tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return fmt.Errorf("validator-join tx failed, tx response code: %v", txRes.Code)

		}
	}

	return nil
}

func (sp *StakingProcessor) sendUnstakeInitToIris(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoUnstakeInit)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.StakingEvent, event); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to iris as already processed",
				"event", eventName,
				"validator", event.User,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"deactivatonEpoch", event.DeactivationEpoch,
				"amount", event.Amount,
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		validNonce, nonceDelay, err := sp.checkValidNonce(event.ValidatorId.Uint64(), event.Nonce.Uint64())
		if err != nil {
			sp.Logger.Error("Error while validating nonce for the validator", "error", err)
			return err
		}

		if nonceDelay > math.MaxInt64 {
			return errors.New("nonceDelay is invalid")
		}

		if !validNonce {
			sp.Logger.Info("Ignoring task to send unstake-init to iris as nonce is out of order")
			//nolint:gosec
			return tasks.NewErrRetryTaskLater("Nonce out of order", defaultDelayDuration*time.Duration(nonceDelay))
		}

		sp.Logger.Info(
			"✅ Received task to send unstake-init to iris",
			"event", eventName,
			"validator", event.User,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"deactivatonEpoch", event.DeactivationEpoch,
			"amount", event.Amount,
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator exit
		msg := stakingTypes.NewMsgValidatorExit(
			hmTypes.BytesToIrisAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			event.DeactivationEpoch.Uint64(),
			hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to iris
		txRes, err := sp.txBroadcaster.BroadcastToIris(msg, event)
		if err != nil {
			sp.Logger.Error("Error while broadcasting unstakeInit to iris", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			sp.Logger.Error("unstakeInit tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return fmt.Errorf("unstakeInit tx failed, tx response code: %v", txRes.Code)

		}
	}

	return nil
}

func (sp *StakingProcessor) sendStakeUpdateToIris(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoStakeUpdate)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.StakingEvent, event); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to iris as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"newAmount", event.NewAmount,
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		validNonce, nonceDelay, err := sp.checkValidNonce(event.ValidatorId.Uint64(), event.Nonce.Uint64())
		if err != nil {
			sp.Logger.Error("Error while validating nonce for the validator", "error", err)
			return err
		}

		if nonceDelay > math.MaxInt64 {
			return errors.New("nonceDelay is invalid")
		}

		if !validNonce {
			sp.Logger.Info("Ignoring task to send stake-update to iris as nonce is out of order")
			//nolint:gosec
			return tasks.NewErrRetryTaskLater("Nonce out of order", defaultDelayDuration*time.Duration(nonceDelay))
		}

		sp.Logger.Info(
			"✅ Received task to send stake-update to iris",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"newAmount", event.NewAmount,
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// msg validator update
		msg := stakingTypes.NewMsgStakeUpdate(
			hmTypes.BytesToIrisAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			sdk.NewIntFromBigInt(event.NewAmount),
			hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to iris
		txRes, err := sp.txBroadcaster.BroadcastToIris(msg, event)
		if err != nil {
			sp.Logger.Error("Error while broadcasting stakeupdate to iris", "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			sp.Logger.Error("stakeupdate tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return fmt.Errorf("stakeupdate tx failed, tx response code: %v", txRes.Code)
		}
	}

	return nil
}

func (sp *StakingProcessor) sendSignerChangeToIris(eventName string, logBytes string) error {
	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		sp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	event := new(stakinginfo.StakinginfoSignerChange)
	if err := helper.UnpackLog(sp.stakingInfoAbi, event, eventName, &vLog); err != nil {
		sp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		newSignerPubKey := event.SignerPubkey
		if len(newSignerPubKey) == 64 {
			newSignerPubKey = util.AppendPrefix(newSignerPubKey)
		}

		if isOld, _ := sp.isOldTx(sp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.StakingEvent, event); isOld {
			sp.Logger.Info("Ignoring task to send unstakeinit to iris as already processed",
				"event", eventName,
				"validatorID", event.ValidatorId,
				"nonce", event.Nonce,
				"NewSignerPubkey", hmTypes.NewPubKey(newSignerPubKey).String(),
				"oldSigner", event.OldSigner.Hex(),
				"newSigner", event.NewSigner.Hex(),
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)
			return nil
		}

		validNonce, nonceDelay, err := sp.checkValidNonce(event.ValidatorId.Uint64(), event.Nonce.Uint64())
		if err != nil {
			sp.Logger.Error("Error while validating nonce for the validator", "error", err)
			return err
		}

		if nonceDelay > math.MaxInt64 {
			return errors.New("nonceDelay is invalid")
		}

		if !validNonce {
			sp.Logger.Info("Ignoring task to send signer-change to iris as nonce is out of order")
			//nolint:gosec
			return tasks.NewErrRetryTaskLater("Nonce out of order", defaultDelayDuration*time.Duration(nonceDelay))
		}

		sp.Logger.Info(
			"✅ Received task to send signer-change to iris",
			"event", eventName,
			"validatorID", event.ValidatorId,
			"nonce", event.Nonce,
			"NewSignerPubkey", hmTypes.NewPubKey(newSignerPubKey).String(),
			"oldSigner", event.OldSigner.Hex(),
			"newSigner", event.NewSigner.Hex(),
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		// signer change
		msg := stakingTypes.NewMsgSignerUpdate(
			hmTypes.BytesToIrisAddress(helper.GetAddress()),
			event.ValidatorId.Uint64(),
			hmTypes.NewPubKey(newSignerPubKey),
			hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Nonce.Uint64(),
		)

		// return broadcast to iris
		txRes, err := sp.txBroadcaster.BroadcastToIris(msg, event)
		if err != nil {
			sp.Logger.Error("Error while broadcasting signerChainge to iris", "msg", msg, "validatorId", event.ValidatorId.Uint64(), "error", err)
			return err
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			sp.Logger.Error("signerChange tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return fmt.Errorf("signerChange tx failed, tx response code: %v", txRes.Code)
		}
	}

	return nil
}

func (sp *StakingProcessor) checkValidNonce(validatorId uint64, txnNonce uint64) (bool, uint64, error) {
	currentNonce, currentHeight, err := util.GetValidatorNonce(sp.cliCtx, validatorId)
	if err != nil {
		sp.Logger.Error("Failed to fetch validator nonce and height data from API", "validatorId", validatorId)
		return false, 0, err
	}

	if currentNonce+1 != txnNonce {
		diff := txnNonce - currentNonce
		if diff > 10 {
			diff = 10
		}

		sp.Logger.Error("Nonce for the given event not in order", "validatorId", validatorId, "currentNonce", currentNonce, "txnNonce", txnNonce, "delay", diff*uint64(defaultDelayDuration))

		return false, diff, nil
	}

	stakingTxnCount, err := queryTxCount(sp.cliCtx, validatorId, currentHeight)
	if err != nil {
		sp.Logger.Error("Failed to query stake txns by txquery for the given validator", "validatorId", validatorId)
		return false, 0, err
	}

	if stakingTxnCount != 0 {
		sp.Logger.Info("Recent staking txn count for the given validator is not zero", "validatorId", validatorId, "currentNonce", currentNonce, "txnNonce", txnNonce, "currentHeight", currentHeight)
		return false, 1, nil
	}

	return true, 0, nil
}

func queryTxCount(cliCtx cliContext.CLIContext, validatorId uint64, currentHeight int64) (int, error) {
	const (
		defaultPage  = 1
		defaultLimit = 30 // should be consistent with tendermint/tendermint/rpc/core/pipe.go:19
	)

	stakingTxnMsgMap := map[string]string{
		"validator-stake-update": "stake-update",
		"validator-join":         "validator-join",
		"signer-update":          "signer-update",
		"validator-exit":         "validator-exit",
	}

	for msg, action := range stakingTxnMsgMap {
		events := []string{
			fmt.Sprintf("%s.%s='%s'", sdk.EventTypeMessage, sdk.AttributeKeyAction, msg),
			fmt.Sprintf("%s.%s=%d", action, "validator-id", validatorId),
			fmt.Sprintf("%s.%s>%d", "tx", "height", currentHeight-3),
		}

		searchResult, err := helper.QueryTxsByEvents(cliCtx, events, defaultPage, defaultLimit)
		if err != nil {
			return 0, err
		}

		if searchResult.TotalCount != 0 {
			return searchResult.TotalCount, nil
		}
	}

	return 0, nil
}
