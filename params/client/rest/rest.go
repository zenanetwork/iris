package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"

	restClient "github.com/zenanetwork/iris/client/rest"
	govRest "github.com/zenanetwork/iris/gov/client/rest"
	govTypes "github.com/zenanetwork/iris/gov/types"
	paramsUtils "github.com/zenanetwork/iris/params/client/utils"
	paramsTypes "github.com/zenanetwork/iris/params/types"
	"github.com/zenanetwork/iris/types/rest"
)

// ProposalRESTHandler returns a ProposalRESTHandler that exposes the param
// change REST handler with a given sub-route.
func ProposalRESTHandler(cliCtx context.CLIContext) govRest.ProposalRESTHandler {
	return govRest.ProposalRESTHandler{
		SubRoute: "param_change",
		Handler:  postProposalHandlerFn(cliCtx),
	}
}

func postProposalHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req paramsUtils.ParamChangeProposalReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		content := paramsTypes.NewParameterChangeProposal(req.Title, req.Description, req.Changes.ToParamChanges())

		msg := govTypes.NewMsgSubmitProposal(content, req.Deposit, req.Proposer, req.Validator)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		restClient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
