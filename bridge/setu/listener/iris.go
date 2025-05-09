package listener

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/RichardKnop/machinery/v1/tasks"
	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"

	checkpointTypes "github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/helper"
	slashingTypes "github.com/zenanetwork/iris/slashing/types"
)

const (
	irisLastBlockKey = "iris-last-block" // storage key
)

// IrisListener - Listens to and process events from iris
type IrisListener struct {
	BaseListener
}

// NewIrisListener - constructor func
func NewIrisListener() *IrisListener {
	return &IrisListener{}
}

// Start starts new block subscription
func (hl *IrisListener) Start() error {
	hl.Logger.Info("Starting")

	// create cancellable context
	headerCtx, cancelHeaderProcess := context.WithCancel(context.Background())
	hl.cancelHeaderProcess = cancelHeaderProcess

	// Iris pollInterval = (minimal pollInterval of rootchain and matichain)
	pollInterval := helper.GetConfig().SyncerPollInterval
	if helper.GetConfig().CheckpointerPollInterval < helper.GetConfig().SyncerPollInterval {
		pollInterval = helper.GetConfig().CheckpointerPollInterval
	}

	hl.Logger.Info("Start polling for events", "pollInterval", pollInterval)
	hl.StartPolling(headerCtx, pollInterval, nil)

	return nil
}

// ProcessHeader -
func (hl *IrisListener) ProcessHeader(_ *blockHeader) {

}

// StartPolling - starts polling for iris events
func (hl *IrisListener) StartPolling(ctx context.Context, pollInterval time.Duration, _ *big.Int) {
	// How often to fire the passed in function in second
	interval := pollInterval

	// Setup the ticket and the channel to signal
	// the ending of the interval
	ticker := time.NewTicker(interval)

	// var eventTypes []string
	// eventTypes = append(eventTypes, "message.action='checkpoint'")
	// eventTypes = append(eventTypes, "message.action='event-record'")
	// eventTypes = append(eventTypes, "message.action='tick'")
	// ADD EVENT TYPE for SLASH-LIMIT

	// start listening
	for {
		select {
		case <-ticker.C:
			fromBlock, toBlock, err := hl.fetchFromAndToBlock()
			if err != nil {
				hl.Logger.Error("Error fetching from and toBlock, skipping events query", "fromBlock", fromBlock, "toBlock", toBlock, "error", err)
			} else if fromBlock < toBlock {

				hl.Logger.Info("Fetching new events between", "fromBlock", fromBlock, "toBlock", toBlock)

				// Querying and processing Begin events
				for i := fromBlock; i <= toBlock; i++ {
					if i > math.MaxInt64 {
						hl.Logger.Error("Block number out of range for int64", "blockNumber", i)
						continue
					}
					//nolint:contextcheck
					events, err := helper.GetBeginBlockEvents(hl.httpClient, int64(i))
					if err != nil {
						hl.Logger.Error("Error fetching begin block events", "error", err)
					}
					for _, event := range events {
						//nolint:gosec
						hl.ProcessBlockEvent(sdk.StringifyEvent(event), int64(i))
					}
				}

				// Querying and processing tx Events. Below for loop is kept for future purpose to process events from tx
				/* 		for _, eventType := range eventTypes {
					var query []string
					query = append(query, eventType)
					query = append(query, fmt.Sprintf("tx.height>=%v", fromBlock))
					query = append(query, fmt.Sprintf("tx.height<=%v", toBlock))

					limit := 50
					for page := 1; page > 0; {
						searchResult, err := helper.QueryTxsByEvents(hl.cliCtx, query, page, limit)
						hl.Logger.Debug("Fetching new events using search query", "query", query, "page", page, "limit", limit)

						if err != nil {
							hl.Logger.Error("Error while searching events", "eventType", eventType, "error", err)
							break
						}

						for _, tx := range searchResult.Txs {
							for _, log := range tx.Logs {
								event := helper.FilterEvents(log.Events, func(et sdk.StringEvent) bool {
									return et.Type == checkpointTypes.EventTypeCheckpoint || et.Type == clerkTypes.EventTypeRecord
								})
								if event != nil {
									hl.ProcessEvent(*event, tx)
								}
							}
						}

						if len(searchResult.Txs) == limit {
							page = page + 1
						} else {
							page = 0
						}
					}
				} */
				// set last block to storage
				if err := hl.storageClient.Put([]byte(irisLastBlockKey), []byte(strconv.FormatUint(toBlock, 10)), nil); err != nil {
					hl.Logger.Error("hl.storageClient.Put", "Error", err)
				}
			}

		case <-ctx.Done():
			hl.Logger.Info("Polling stopped")
			ticker.Stop()

			return
		}
	}
}

func (hl *IrisListener) fetchFromAndToBlock() (uint64, uint64, error) {
	// toBlock - get latest blockheight from iris node
	fromBlock := uint64(0)
	toBlock := uint64(0)

	nodeStatus, err := helper.GetNodeStatus(hl.cliCtx)
	if err != nil {
		hl.Logger.Error("Error while fetching iris node status", "error", err)
		return fromBlock, toBlock, err
	}
	if nodeStatus.SyncInfo.LatestBlockHeight < 0 {
		return fromBlock, toBlock, fmt.Errorf("latest block height is negative: %d", nodeStatus.SyncInfo.LatestBlockHeight)
	}

	toBlock = big.NewInt(0).SetInt64(nodeStatus.SyncInfo.LatestBlockHeight).Uint64()

	// fromBlock - get last block from storage
	hasLastBlock, _ := hl.storageClient.Has([]byte(irisLastBlockKey), nil)
	if hasLastBlock {
		lastBlockBytes, e := hl.storageClient.Get([]byte(irisLastBlockKey), nil)
		if e != nil {
			hl.Logger.Info("Error while fetching last block bytes from storage", "error", e)
			return fromBlock, toBlock, e
		}

		//nolint:gosec
		if result, e := strconv.ParseUint(string(lastBlockBytes), 10, 64); e == nil {
			hl.Logger.Debug("Got last block from bridge storage", "lastBlock", result)
			fromBlock = result + 1
		} else {
			hl.Logger.Info("Error parsing last block bytes from storage", "error", err)
			toBlock = 0

			return fromBlock, toBlock, e
		}
	}

	return fromBlock, toBlock, nil
}

// ProcessBlockEvent - process Blockevents (BeginBlock, EndBlock events) from iris.
func (hl *IrisListener) ProcessBlockEvent(event sdk.StringEvent, blockHeight int64) {
	hl.Logger.Info("Received block event from Iris", "eventType", event.Type)

	eventBytes, err := jsoniter.ConfigFastest.Marshal(event)
	if err != nil {
		hl.Logger.Error("Error while parsing block event", "eventType", event.Type, "error", err)
		return
	}

	switch event.Type {
	case checkpointTypes.EventTypeCheckpoint:
		hl.sendBlockTask("sendCheckpointToRootchain", eventBytes, blockHeight)
	case slashingTypes.EventTypeSlashLimit:
		hl.sendBlockTask("sendTickToIris", eventBytes, blockHeight)
	case slashingTypes.EventTypeTickConfirm:
		hl.sendBlockTask("sendTickToRootchain", eventBytes, blockHeight)
	default:
		hl.Logger.Debug("BlockEvent Type mismatch", "eventType", event.Type)
	}
}

func (hl *IrisListener) sendBlockTask(taskName string, eventBytes []byte, blockHeight int64) {
	// create machinery task
	signature := &tasks.Signature{
		Name: taskName,
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: string(eventBytes),
			},
			{
				Type:  "int64",
				Value: blockHeight,
			},
		},
	}
	signature.RetryCount = 3

	hl.Logger.Info("Sending block level task", "taskName", taskName, "currentTime", time.Now(), "blockHeight", blockHeight)

	// send task
	_, err := hl.queueConnector.Server.SendTask(signature)
	if err != nil {
		hl.Logger.Error("Error sending block level task", "taskName", taskName, "blockHeight", blockHeight, "error", err)
	}
}
