// nolint
package rest

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/iris/zena/types"
	restClient "github.com/zenanetwork/iris/client/rest"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/types/rest"
)

// It represents Propose Span msg.
//
//swagger:response zenaProposeSpanResponse
type zenaProposeSpanResponse struct {
	//in:body
	Output output `json:"output"`
}

type output struct {
	Type  string `json:"type"`
	Value value  `json:"value"`
}

type value struct {
	Msg       msg    `json:"msg"`
	Signature string `json:"signature"`
	Memo      string `json:"memo"`
}

type msg struct {
	Type  string `json:"type"`
	Value val    `json:"value"`
}

type val struct {
	SpanID     string `json:"span_id"`
	Proposer   string `json:"proposer"`
	StartBlock string `json:"start_block"`
	EndBlock   string `json:"end_block"`
	ZenaChainId string `json:"zena_chain_id"`
	Seed       string `json:"seed"`
}

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/zena/propose-span",
		postProposeSpanHandlerFn(cliCtx),
	).Methods("POST")
}

// ProposeSpanReq struct for proposing new span
type ProposeSpanReq struct {
	BaseReq rest.BaseReq `json:"base_req"`

	ID         uint64 `json:"span_id"`
	StartBlock uint64 `json:"start_block"`
	ZenaChainID string `json:"zena_chain_id"`
}

//swagger:parameters zenaProposeSpan
type zenaProposeSpan struct {

	//Body
	//required:true
	//in:body
	Input SendReqInput `json:"input"`
}

type SendReqInput struct {

	//required:true
	//in:body
	BaseReq BaseReq `json:"base_req"`

	//required:true
	//in:body
	ID uint64 `json:"span_id"`

	//required:true
	//in:body
	StartBlock uint64 `json:"start_block"`

	//required:true
	//in:body
	ZenaChainID string `json:"zena_chain_id"`
}

type BaseReq struct {

	//Address of the sender
	//required:true
	//in:body
	From string `json:"address"`

	//Chain ID of Iris
	//required:true
	//in:body
	ChainID string `json:"chain_id"`
}

// swagger:route POST /zena/propose-span zena zenaProposeSpan
// It returns the prepared msg for proposing the span
// responses:
//   200: zenaProposeSpanResponse

func postProposeSpanHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read req from request
		var req ProposeSpanReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		//
		// Get span duration
		//

		// fetch duration
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryParams, types.ParamSpan), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if len(res) == 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, errors.New("Span duration not found ").Error())
			return
		}

		var spanDuration uint64
		if err = jsoniter.ConfigFastest.Unmarshal(res, &spanDuration); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		seedQueryParams, err := cliCtx.Codec.MarshalJSON(types.NewQuerySpanParams(req.ID))
		if err != nil {
			return
		}

		// fetch seed
		res, _, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryNextSpanSeed), seedQueryParams)
		if err != nil {
			RestLogger.Error("Error while fetching next span seed  ", "Error", err.Error())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		var seedResponse types.QuerySpanSeedResponse
		if err = jsoniter.ConfigFastest.Unmarshal(res, &seedResponse); err != nil {
			return
		}

		nodeStatus, err := helper.GetNodeStatus(cliCtx)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var msg sdk.Msg
		if nodeStatus.SyncInfo.LatestBlockHeight < helper.GetDanelawHeight() {
			// draft a propose span message
			msg = types.NewMsgProposeSpan(
				req.ID,
				hmTypes.HexToIrisAddress(req.BaseReq.From),
				req.StartBlock,
				req.StartBlock+spanDuration-1,
				req.ZenaChainID,
				seedResponse.Seed,
			)
		} else {
			// draft a propose span v2 message
			msg = types.NewMsgProposeSpanV2(
				req.ID,
				hmTypes.HexToIrisAddress(req.BaseReq.From),
				req.StartBlock,
				req.StartBlock+spanDuration-1,
				req.ZenaChainID,
				seedResponse.Seed,
				seedResponse.SeedAuthor,
			)
		}

		// send response
		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
