// Package classification HiemdallRest API
//
//	    Schemes: http
//	    BasePath: /
//	    Version: 0.0.1
//	    title: Iris APIs
//	    Consumes:
//	    - application/json
//		   Host:localhost:1317
//	    - application/json
//
// nolint
//
//swagger:meta
package rest

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"

	"github.com/zenanetwork/iris/zena/types"
	checkpointTypes "github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/helper"
	stakingTypes "github.com/zenanetwork/iris/staking/types"
	hmTypes "github.com/zenanetwork/iris/types"
	hmRest "github.com/zenanetwork/iris/types/rest"
)

type IrisSpanResultWithHeight struct {
	Height int64
	Result []byte
}

type validator struct {
	ID           int    `json:"ID"`
	StartEpoch   int    `json:"startEpoch"`
	EndEpoch     int    `json:"endEpoch"`
	Nonce        int    `json:"nonce"`
	Power        int    `json:"power"`
	PubKey       string `json:"pubKey"`
	Signer       string `json:"signer"`
	Last_Updated string `json:"last_updated"`
	Jailed       bool   `json:"jailed"`
	Accum        int    `json:"accum"`
}

type span struct {
	SpanID     int `json:"span_id"`
	StartBlock int `json:"start_block"`
	EndBlock   int `json:"end_block"`
	//in:body
	ValidatorSet      validatorSet `json:"validator_set"`
	SelectedProducers []validator  `json:"selected_producer"`
	ZenaChainId        string       `json:"zena_chain_id"`
}

type validatorSet struct {
	Validators []validator `json:"validators"`
	Proposer   validator   `json:"Proposer"`
}

// It represents the list of spans
//
//swagger:response zenaSpanListResponse
type zenaSpanListResponse struct {
	//in:body
	Output zenaSpanList `json:"output"`
}

type zenaSpanList struct {
	Height string `json:"height"`
	Result []span `json:"result"`
}

// It represents the span
//
//swagger:response zenaSpanResponse
type zenaSpanResponse struct {
	//in:body
	Output zenaSpan `json:"output"`
}

type zenaSpan struct {
	Height string `json:"height"`
	Result span   `json:"result"`
}

// It represents the zena span parameters
//
//swagger:response zenaSpanParamsResponse
type zenaSpanParamsResponse struct {
	//in:body
	Output zenaSpanParams `json:"output"`
}

type zenaSpanParams struct {
	Height string     `json:"height"`
	Result spanParams `json:"result"`
}

type spanParams struct {

	//type:integer
	SprintDuration int64 `json:"sprint_duration"`
	//type:integer
	SpanDuration int64 `json:"span_duration"`
	//type:integer
	ProducerCount int64 `json:"producer_count"`
}

// It represents the next span seed
//
//swagger:response zenaNextSpanSeedResponse
type zenaNextSpanSeedResponse struct {
	//in:body
	Output spanSeed `json:"output"`
}

type spanSeed struct {
	Height string `json:"height"`
	Result string `json:"result"`
}

var spanOverrides map[uint64]*IrisSpanResultWithHeight = nil

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/zena/span/list", spanListHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/zena/span/{id}", spanHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/zena/latest-span", latestSpanHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/zena/prepare-next-span", prepareNextSpanHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/zena/next-span-seed/{id}", fetchNextSpanSeedHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/zena/params", paramsHandlerFn(cliCtx)).Methods("GET")
}

//swagger:parameters zenaCurrentSpanById
type zenaCurrentSpanById struct {

	//Id number of the span
	//required:true
	//type:integer
	//in:path
	Id int `json:"id"`
}

// swagger:route GET /zena/next-span-seed/{id} zena zenaCurrentSpanById
// It returns the seed for the next span
// responses:
//   200: zenaNextSpanSeedResponse

func fetchNextSpanSeedHandlerFn(
	cliCtx context.CLIContext,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		spanID, ok := rest.ParseUint64OrReturnBadRequest(w, vars["id"])
		if !ok {
			return
		}

		seedQueryParams, err := cliCtx.Codec.MarshalJSON(types.NewQuerySpanParams(spanID))
		if err != nil {
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryNextSpanSeed), seedQueryParams)
		if err != nil {
			RestLogger.Error("Error while fetching next span seed  ", "Error", err.Error())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		RestLogger.Debug("nextSpanSeed querier response", "res", res)

		// error if span seed found
		if !hmRest.ReturnNotFoundIfNoContent(w, res, "NextSpanSeed not found") {
			RestLogger.Error("NextSpanSeed not found ", "Error", err)
			return
		}

		// return result
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

//swagger:parameters zenaSpanList
type zenaSpanListParam struct {

	//Page Number
	//required:true
	//type:integer
	//in:query
	Page int `json:"page"`

	//Limit
	//required:true
	//type:integer
	//in:query
	Limit int `json:"limit"`
}

// swagger:route GET /zena/span/list zena zenaSpanList
// It returns the list of Zena Span
// responses:
//
//	200: zenaSpanListResponse
func spanListHandlerFn(
	cliCtx context.CLIContext,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// get page
		page, ok := rest.ParseUint64OrReturnBadRequest(w, vars.Get("page"))
		if !ok {
			return
		}

		// get limit
		limit, ok := rest.ParseUint64OrReturnBadRequest(w, vars.Get("limit"))
		if !ok {
			return
		}

		// get query params
		queryParams, err := cliCtx.Codec.MarshalJSON(hmTypes.NewQueryPaginationParams(page, limit))
		if err != nil {
			return
		}

		// query spans
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySpanList), queryParams)
		if err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, res, "No spans found"); !ok {
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

//swagger:parameters zenaSpanById
type zenaSpanById struct {

	//Id number of the span
	//required:true
	//type:integer
	//in:path
	Id int `json:"id"`
}

// swagger:route GET /zena/span/{id} zena zenaSpanById
// It returns the span based on ID
// responses:
//
//	200: zenaSpanResponse
func spanHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		vars := mux.Vars(r)

		// get to address
		spanID, ok := rest.ParseUint64OrReturnBadRequest(w, vars["id"])
		if !ok {
			return
		}

		var (
			res            []byte
			height         int64
			spanOverridden bool
		)

		if spanOverrides == nil {
			loadSpanOverrides()
		}

		if span, ok := spanOverrides[spanID]; ok {
			res = span.Result
			height = span.Height
			spanOverridden = true
		}

		if !spanOverridden {
			// get query params
			queryParams, err := cliCtx.Codec.MarshalJSON(types.NewQuerySpanParams(spanID))
			if err != nil {
				return
			}

			// fetch span
			res, height, err = cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QuerySpan), queryParams)
			if err != nil {
				hmRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, res, "No span found"); !ok {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		hmRest.PostProcessResponse(w, cliCtx, res)
	}
}

// swagger:route GET /zena/latest-span zena zenaSpanLatest
// It returns the latest-span
// responses:
//
//	200: zenaSpanResponse
func latestSpanHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// fetch latest span
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryLatestSpan), nil)
		if err != nil {
			hmRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, res, "No latest span found"); !ok {
			return
		}

		// return result
		cliCtx = cliCtx.WithHeight(height)
		hmRest.PostProcessResponse(w, cliCtx, res)
	}
}

//swagger:parameters zenaPrepareNextSpan
type zenaPrepareNextSpanParam struct {

	//Start Block
	//required:true
	//type:integer
	//in:query
	StartBlock int `json:"start_block"`

	//Span ID of the span
	//required:true
	//type:integer
	//in:query
	SpanId int `json:"span_id"`

	//Chain ID of the network
	//required:true
	//type:integer
	//in:query
	ChainId int `json:"chain_id"`
}

// swagger:route GET /zena/prepare-next-span zena zenaPrepareNextSpan
// It returns the prepared next span
// responses:
//
//	200: zenaSpanResponse
func prepareNextSpanHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		params := r.URL.Query()

		spanID, ok := rest.ParseUint64OrReturnBadRequest(w, params.Get("span_id"))
		if !ok {
			return
		}

		startBlock, ok := rest.ParseUint64OrReturnBadRequest(w, params.Get("start_block"))
		if !ok {
			return
		}

		chainID := params.Get("chain_id")

		//
		// Get span duration
		//

		// fetch duration
		spanDurationBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", types.QuerierRoute, types.QueryParams, types.ParamSpan), nil)
		if err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, spanDurationBytes, "No span duration"); !ok {
			return
		}

		var spanDuration uint64
		if err := jsoniter.ConfigFastest.Unmarshal(spanDurationBytes, &spanDuration); err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		//
		// Get ack count
		//

		// fetch ack count
		ackCountBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", checkpointTypes.QuerierRoute, checkpointTypes.QueryAckCount), nil)
		if err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, ackCountBytes, "Ack not found"); !ok {
			return
		}

		var ackCount uint64
		if err := jsoniter.ConfigFastest.Unmarshal(ackCountBytes, &ackCount); err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		//
		// Validators
		//

		validatorSetBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", stakingTypes.QuerierRoute, stakingTypes.QueryCurrentValidatorSet), nil)
		if err != nil {
			hmRest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// check content
		if !hmRest.ReturnNotFoundIfNoContent(w, validatorSetBytes, "No current validator set found") {
			return
		}

		var _validatorSet hmTypes.ValidatorSet
		if err = jsoniter.ConfigFastest.Unmarshal(validatorSetBytes, &_validatorSet); err != nil {
			hmRest.WriteErrorResponse(w, http.StatusNoContent, errors.New("unable to unmarshall JSON").Error())
			return
		}

		//
		// Fetching SelectedProducers
		//

		query, err := cliCtx.Codec.MarshalJSON(types.NewQuerySpanParams(spanID))
		if err != nil {
			fmt.Println("error while marshalling: ", err)
			hmRest.WriteErrorResponse(w, http.StatusNoContent, errors.New("unable to marshal JSON").Error())
			return
		}

		nextProducerBytes, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryNextProducers), query)
		if err != nil {
			fmt.Println("error while querying next producers: ", err)
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// check content
		if ok := hmRest.ReturnNotFoundIfNoContent(w, nextProducerBytes, "Next Producers not found"); !ok {
			return
		}

		var selectedProducers []hmTypes.Validator
		if err := jsoniter.ConfigFastest.Unmarshal(nextProducerBytes, &selectedProducers); err != nil {
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		selectedProducers = hmTypes.SortValidatorByAddress(selectedProducers)

		// draft a propose span message
		msg := hmTypes.NewSpan(
			spanID,
			startBlock,
			startBlock+spanDuration-1,
			_validatorSet,
			selectedProducers,
			chainID,
		)

		result, err := jsoniter.ConfigFastest.Marshal(&msg)
		if err != nil {
			RestLogger.Error("Error while marshalling response to Json", "error", err)
			hmRest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())

			return
		}

		hmRest.PostProcessResponse(w, cliCtx, result)
	}
}

// swagger:route GET /zena/params zena zenaSpanParams
// It returns the span parameters
// responses:
//
//	200: zenaSpanParamsResponse
func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ResponseWithHeight defines a response object type that wraps an original
// response with a height.
// TODO:Link it with zena
type ResponseWithHeight struct {
	Height string              `json:"height"`
	Result jsoniter.RawMessage `json:"result"`
}

func loadSpanOverrides() {
	spanOverrides = map[uint64]*IrisSpanResultWithHeight{}

	j, ok := SPAN_OVERRIDES[helper.GenesisDoc.ChainID]
	if !ok {
		return
	}

	var spans []*types.ResponseWithHeight
	if err := jsoniter.ConfigFastest.Unmarshal(j, &spans); err != nil {
		return
	}

	for _, span := range spans {
		var irisSpan types.IrisSpan
		if err := jsoniter.ConfigFastest.Unmarshal(span.Result, &irisSpan); err != nil {
			continue
		}

		height, err := strconv.ParseInt(span.Height, 10, 64)
		if err != nil {
			continue
		}

		spanOverrides[irisSpan.ID] = &IrisSpanResultWithHeight{
			Height: height,
			Result: span.Result,
		}
	}
}

//swagger:parameters zenaSpanList zenaSpanById zenaPrepareNextSpan zenaSpanLatest zenaSpanParams zenaNextSpanSeed
type Height struct {

	//Block Height
	//in:query
	Height string `json:"height"`
}
