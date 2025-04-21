package processor

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
	jsoniter "github.com/json-iterator/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/zenanetwork/go-zenanet/accounts/abi"
	"github.com/zenanetwork/go-zenanet/core/types"

	"github.com/zenanetwork/iris/bridge/setu/util"
	chainmanagerTypes "github.com/zenanetwork/iris/chainmanager/types"
	clerkTypes "github.com/zenanetwork/iris/clerk/types"
	"github.com/zenanetwork/iris/common/tracing"
	"github.com/zenanetwork/iris/contracts/statesender"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
)

// ClerkContext for bridge
type ClerkContext struct {
	ChainmanagerParams *chainmanagerTypes.Params
}

// ClerkProcessor - sync state/deposit events
type ClerkProcessor struct {
	BaseProcessor
	stateSenderAbi *abi.ABI
}

// NewClerkProcessor - add statesender abi to clerk processor
func NewClerkProcessor(stateSenderAbi *abi.ABI) *ClerkProcessor {
	return &ClerkProcessor{
		stateSenderAbi: stateSenderAbi,
	}
}

// Start starts new block subscription
func (cp *ClerkProcessor) Start() error {
	cp.Logger.Info("Starting")
	return nil
}

// RegisterTasks - Registers clerk related tasks with machinery
func (cp *ClerkProcessor) RegisterTasks() {
	cp.Logger.Info("Registering clerk tasks")

	if err := cp.queueConnector.Server.RegisterTask("sendStateSyncedToIris", cp.sendStateSyncedToIris); err != nil {
		cp.Logger.Error("RegisterTasks | sendStateSyncedToIris", "error", err)
	}
}

// HandleStateSyncEvent - handle state sync event from rootchain
// 1. check if this deposit event has to be broadcasted to iris
// 2. create and broadcast  record transaction to iris
func (cp *ClerkProcessor) sendStateSyncedToIris(eventName string, logBytes string) error {
	otelCtx := tracing.WithTracer(context.Background(), otel.Tracer("State-Sync"))
	// work begins
	sendStateSyncedToIrisCtx, sendStateSyncedToIrisSpan := tracing.StartSpan(otelCtx, "sendStateSyncedToIris")
	defer tracing.EndSpan(sendStateSyncedToIrisSpan)

	start := time.Now()

	var vLog = types.Log{}
	if err := jsoniter.ConfigFastest.Unmarshal([]byte(logBytes), &vLog); err != nil {
		cp.Logger.Error("Error while unmarshalling event from rootchain", "error", err)
		return err
	}

	clerkContext, err := cp.getClerkContext()
	if err != nil {
		return err
	}

	chainParams := clerkContext.ChainmanagerParams.ChainParams

	event := new(statesender.StatesenderStateSynced)
	if err = helper.UnpackLog(cp.stateSenderAbi, event, eventName, &vLog); err != nil {
		cp.Logger.Error("Error while parsing event", "name", eventName, "error", err)
	} else {
		defer util.LogElapsedTimeForStateSyncedEvent(event, "sendStateSyncedToIris", start)

		tracing.SetAttributes(sendStateSyncedToIrisSpan, []attribute.KeyValue{
			attribute.String("event", eventName),
			attribute.Int64("id", event.Id.Int64()),
			attribute.String("contract", event.ContractAddress.String()),
		}...)

		_, isOldTxSpan := tracing.StartSpan(sendStateSyncedToIrisCtx, "isOldTx")
		isOld, _ := cp.isOldTx(cp.cliCtx, vLog.TxHash.String(), uint64(vLog.Index), util.ClerkEvent, event)
		tracing.EndSpan(isOldTxSpan)

		if isOld {
			cp.Logger.Info("Ignoring task to send deposit to iris as already processed",
				"event", eventName,
				"id", event.Id,
				"contract", event.ContractAddress,
				"data", hex.EncodeToString(event.Data),
				"zenaChainId", chainParams.ZenaChainID,
				"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
				"logIndex", uint64(vLog.Index),
				"blockNumber", vLog.BlockNumber,
			)

			return nil
		}

		cp.Logger.Debug(
			"⬜ New event found",
			"event", eventName,
			"id", event.Id,
			"contract", event.ContractAddress,
			"data", hex.EncodeToString(event.Data),
			"zenaChainId", chainParams.ZenaChainID,
			"txHash", hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			"logIndex", uint64(vLog.Index),
			"blockNumber", vLog.BlockNumber,
		)

		_, maxStateSyncSizeCheckSpan := tracing.StartSpan(sendStateSyncedToIrisCtx, "maxStateSyncSizeCheck")
		if util.GetBlockHeight(cp.cliCtx) > helper.GetSpanOverrideHeight() && len(event.Data) > helper.MaxStateSyncSize {
			cp.Logger.Info(`Data is too large to process, Resetting to ""`, "data", hex.EncodeToString(event.Data))
			event.Data = hmTypes.HexToHexBytes("")
		} else if len(event.Data) > helper.LegacyMaxStateSyncSize {
			cp.Logger.Info(`Data is too large to process, Resetting to ""`, "data", hex.EncodeToString(event.Data))
			event.Data = hmTypes.HexToHexBytes("")
		}
		tracing.EndSpan(maxStateSyncSizeCheckSpan)

		msg := clerkTypes.NewMsgEventRecord(
			hmTypes.BytesToIrisAddress(helper.GetAddress()),
			hmTypes.BytesToIrisHash(vLog.TxHash.Bytes()),
			uint64(vLog.Index),
			vLog.BlockNumber,
			event.Id.Uint64(),
			hmTypes.BytesToIrisAddress(event.ContractAddress.Bytes()),
			event.Data,
			chainParams.ZenaChainID,
		)

		_, checkTxAgainstMempoolSpan := tracing.StartSpan(sendStateSyncedToIrisCtx, "checkTxAgainstMempool")
		// Check if we have the same transaction in mempool or not
		// Don't drop the transaction. Keep retrying after `util.RetryStateSyncTaskDelay = 24 seconds`,
		// until the transaction in mempool is processed or cancelled.
		inMempool, _ := cp.checkTxAgainstMempool(msg, event)
		tracing.EndSpan(checkTxAgainstMempoolSpan)

		if inMempool {
			cp.Logger.Info("Similar transaction already in mempool, retrying in sometime", "event", eventName, "retry delay", util.RetryStateSyncTaskDelay)
			return tasks.NewErrRetryTaskLater("transaction already in mempool", util.RetryStateSyncTaskDelay)
		}

		_, BroadcastToIrisSpan := tracing.StartSpan(sendStateSyncedToIrisCtx, "BroadcastToIris")
		// return broadcast to iris
		_, err = cp.txBroadcaster.BroadcastToIris(msg, event)
		tracing.EndSpan(BroadcastToIrisSpan)

		if err != nil {
			cp.Logger.Error("Error while broadcasting clerk Record to iris", "error", err)
			return err
		}
	}

	return nil
}

//
// utils
//

func (cp *ClerkProcessor) getClerkContext() (*ClerkContext, error) {
	chainmanagerParams, err := util.GetChainmanagerParams(cp.cliCtx)
	if err != nil {
		cp.Logger.Error("Error while fetching chain manager params", "error", err)
		return nil, err
	}

	return &ClerkContext{
		ChainmanagerParams: chainmanagerParams,
	}, nil
}
