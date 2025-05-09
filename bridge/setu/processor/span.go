package processor

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/go-zenanet/common"

	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/helper"
	"github.com/zenanetwork/iris/types"
	zenaTypes "github.com/zenanetwork/iris/zena/types"
)

// SpanProcessor - process span related events
type SpanProcessor struct {
	BaseProcessor

	// header listener subscription
	cancelSpanService context.CancelFunc
}

// Start starts new block subscription
func (sp *SpanProcessor) Start() error {
	sp.Logger.Info("Starting")

	// create cancellable context
	spanCtx, cancelSpanService := context.WithCancel(context.Background())

	sp.cancelSpanService = cancelSpanService

	// start polling for span
	sp.Logger.Info("Start polling for span", "pollInterval", helper.GetConfig().SpanPollInterval)

	go sp.startPolling(spanCtx, helper.GetConfig().SpanPollInterval)

	return nil
}

// RegisterTasks - nil
func (sp *SpanProcessor) RegisterTasks() {

}

// startPolling - polls iris and checks if new span needs to be proposed
func (sp *SpanProcessor) startPolling(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	// stop ticker when everything done
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// nolint: contextcheck
			sp.checkAndPropose()
		case <-ctx.Done():
			sp.Logger.Info("Polling stopped")
			ticker.Stop()

			return
		}
	}
}

// checkAndPropose - will check if current user is span proposer and proposes the span
func (sp *SpanProcessor) checkAndPropose() {
	lastSpan, err := sp.getLastSpan()
	if err != nil {
		sp.Logger.Error("Unable to fetch last span", "error", err)
		return
	}

	if lastSpan == nil {
		return
	}

	nodeStatus, err := helper.GetNodeStatus(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error while fetching iris node status", "error", err)
		return
	}

	if nodeStatus.SyncInfo.LatestBlockHeight >= helper.GetDanelawHeight() {
		latestBlock, e := sp.contractConnector.GetMaticChainBlock(nil)
		if e != nil {
			sp.Logger.Error("Error fetching current child block", "error", e)
			return
		}

		if latestBlock.Number.Uint64() < lastSpan.StartBlock {
			sp.Logger.Debug("Current zena block is less than last span start block, skipping proposing span", "currentBlock", latestBlock.Number.Uint64(), "lastSpanStartBlock", lastSpan.StartBlock)
			return
		}
	}

	sp.Logger.Debug("Found last span", "lastSpan", lastSpan.ID, "startBlock", lastSpan.StartBlock, "endBlock", lastSpan.EndBlock)

	nextSpanMsg, err := sp.fetchNextSpanDetails(lastSpan.ID+1, lastSpan.EndBlock+1)
	if err != nil {
		sp.Logger.Error("Unable to fetch next span details", "error", err, "lastSpanId", lastSpan.ID)
		return
	}

	// check if current user is among next span producers
	if sp.isSpanProposer(nextSpanMsg.SelectedProducers) {
		go sp.propose(lastSpan, nextSpanMsg)
	}
}

// propose producers for next span if needed
func (sp *SpanProcessor) propose(lastSpan *types.Span, nextSpanMsg *types.Span) {
	// call with last span on record + new span duration and see if it has been proposed
	currentBlock, err := sp.getCurrentChildBlock()
	if err != nil {
		sp.Logger.Error("Unable to fetch current block", "error", err)
		return
	}

	if lastSpan.StartBlock <= currentBlock && currentBlock <= lastSpan.EndBlock {
		// log new span
		sp.Logger.Info("✅ Proposing new span", "spanId", nextSpanMsg.ID, "startBlock", nextSpanMsg.StartBlock, "endBlock", nextSpanMsg.EndBlock)

		seed, seedAuthor, err := sp.fetchNextSpanSeed(nextSpanMsg.ID)
		if err != nil {
			sp.Logger.Info("Error while fetching next span seed from IrisServer", "err", err)
			return
		}

		nodeStatus, err := helper.GetNodeStatus(sp.cliCtx)
		if err != nil {
			sp.Logger.Error("Error while fetching iris node status", "error", err)
			return
		}

		var txRes sdk.TxResponse

		if nodeStatus.SyncInfo.LatestBlockHeight < helper.GetDanelawHeight() {
			// broadcast to iris
			msg := zenaTypes.MsgProposeSpan{
				ID:         nextSpanMsg.ID,
				Proposer:   types.BytesToIrisAddress(helper.GetAddress()),
				StartBlock: nextSpanMsg.StartBlock,
				EndBlock:   nextSpanMsg.EndBlock,
				ChainID:    nextSpanMsg.ChainID,
				Seed:       seed,
			}

			// return broadcast to iris
			txRes, err = sp.txBroadcaster.BroadcastToIris(msg, nil)
			if err != nil {
				sp.Logger.Error("Error while broadcasting span to iris", "spanId", nextSpanMsg.ID, "startBlock", nextSpanMsg.StartBlock, "endBlock", nextSpanMsg.EndBlock, "error", err)
				return
			}
		} else {
			msg := zenaTypes.MsgProposeSpanV2{
				ID:         nextSpanMsg.ID,
				Proposer:   types.BytesToIrisAddress(helper.GetAddress()),
				StartBlock: nextSpanMsg.StartBlock,
				EndBlock:   nextSpanMsg.EndBlock,
				ChainID:    nextSpanMsg.ChainID,
				Seed:       seed,
				SeedAuthor: seedAuthor,
			}

			txRes, err = sp.txBroadcaster.BroadcastToIris(msg, nil)
			if err != nil {
				sp.Logger.Error("Error while broadcasting span to iris", "spanId", nextSpanMsg.ID, "startBlock", nextSpanMsg.StartBlock, "endBlock", nextSpanMsg.EndBlock, "error", err)
				return
			}
		}

		if txRes.Code != uint32(sdk.CodeOK) {
			sp.Logger.Error("span tx failed on iris", "txHash", txRes.TxHash, "code", txRes.Code)
			return

		}
	}
}

// checks span status
func (sp *SpanProcessor) getLastSpan() (*types.Span, error) {
	// fetch latest start block from iris via rest query
	result, err := helper.FetchFromAPI(sp.cliCtx, helper.GetIrisServerEndpoint(util.LatestSpanURL))
	if err != nil {
		sp.Logger.Error("Error while fetching latest span")
		return nil, err
	}

	var lastSpan types.Span
	if err = jsoniter.ConfigFastest.Unmarshal(result.Result, &lastSpan); err != nil {
		sp.Logger.Error("Error unmarshalling span", "error", err)
		return nil, err
	}

	return &lastSpan, nil
}

// getCurrentChildBlock gets the current child block
func (sp *SpanProcessor) getCurrentChildBlock() (uint64, error) {
	childBlock, err := sp.contractConnector.GetMaticChainBlock(nil)
	if err != nil {
		return 0, err
	}

	return childBlock.Number.Uint64(), nil
}

// isSpanProposer checks if current user is span proposer
func (sp *SpanProcessor) isSpanProposer(nextSpanProducers []types.Validator) bool {
	// anyone among next span producers can become next span proposer
	for _, val := range nextSpanProducers {
		if bytes.Equal(val.Signer.Bytes(), helper.GetAddress()) {
			return true
		}
	}

	return false
}

// fetch next span details from iris.
func (sp *SpanProcessor) fetchNextSpanDetails(id uint64, start uint64) (*types.Span, error) {
	req, err := http.NewRequest("GET", helper.GetIrisServerEndpoint(util.NextSpanInfoURL), nil)
	if err != nil {
		sp.Logger.Error("Error creating a new request", "error", err)
		return nil, err
	}

	configParams, err := util.GetChainmanagerParams(sp.cliCtx)
	if err != nil {
		sp.Logger.Error("Error while fetching chainmanager params", "error", err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("span_id", strconv.FormatUint(id, 10))
	q.Add("start_block", strconv.FormatUint(start, 10))
	q.Add("chain_id", configParams.ChainParams.ZenaChainID)
	q.Add("proposer", helper.GetFromAddress(sp.cliCtx).String())
	req.URL.RawQuery = q.Encode()

	// fetch next span details
	result, err := helper.FetchFromAPI(sp.cliCtx, req.URL.String())
	if err != nil {
		sp.Logger.Error("Error fetching proposers", "error", err)
		return nil, err
	}

	var msg types.Span
	if err = jsoniter.ConfigFastest.Unmarshal(result.Result, &msg); err != nil {
		sp.Logger.Error("Error unmarshalling propose tx msg ", "error", err)
		return nil, err
	}

	sp.Logger.Debug("◽ Generated proposer span msg", "msg", msg.String())

	return &msg, nil
}

// fetchNextSpanSeed - fetches seed for next span
func (sp *SpanProcessor) fetchNextSpanSeed(id uint64) (common.Hash, common.Address, error) {
	sp.Logger.Info("Sending Rest call to Get Seed for next span")

	response, err := helper.FetchFromAPI(sp.cliCtx, helper.GetIrisServerEndpoint(fmt.Sprintf(util.NextSpanSeedURL, strconv.FormatUint(id, 10))))
	if err != nil {
		sp.Logger.Error("Error Fetching nextspanseed from IrisServer ", "error", err)
		return common.Hash{}, common.Address{}, err
	}

	sp.Logger.Info("Next span seed fetched")

	var nextSpanSeedResponse zenaTypes.QuerySpanSeedResponse

	if err = jsoniter.ConfigFastest.Unmarshal(response.Result, &nextSpanSeedResponse); err != nil {
		sp.Logger.Error("Error unmarshalling nextSpanSeed received from Iris Server", "error", err)
		return common.Hash{}, common.Address{}, err
	}

	return nextSpanSeedResponse.Seed, nextSpanSeedResponse.SeedAuthor, nil
}

// Stop stops all necessary go routines
func (sp *SpanProcessor) Stop() {
	// cancel span polling
	sp.cancelSpanService()
}
