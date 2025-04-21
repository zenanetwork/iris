package simulation

import (
	"github.com/zenanetwork/iris/bank/types"
	"github.com/zenanetwork/iris/types/module"
)

// RandomizedGenState returns bank genesis
func RandomizedGenState(simState *module.SimulationState) {
	bankGenesis := types.NewGenesisState(true)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(bankGenesis)
}
