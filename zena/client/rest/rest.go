package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/gorilla/mux"
	tmLog "github.com/tendermint/tendermint/libs/log"

	"github.com/zenanetwork/iris/helper"
)

// RestLogger for zena module logger
var RestLogger tmLog.Logger

func init() {
	RestLogger = helper.Logger.With("module", "zena/rest")
}

// RegisterRoutes registers  zena-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}
