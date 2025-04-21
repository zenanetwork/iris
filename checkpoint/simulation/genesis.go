//nolint:gosec
package simulation

import (
	"time"

	"github.com/zenanetwork/iris/checkpoint/types"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/types/module"
)

// RandomizedGenState return dummy genesis
func RandomizedGenState(simState *module.SimulationState) {
	lastNoACK := 0
	ackCount := 1
	startBlock := uint64(0)
	endBlock := uint64(256)
	rootHash := hmTypes.HexToIrisHash("123")

	proposerAddress := hmTypes.HexToIrisAddress("123")
	timestamp := uint64(time.Now().Unix())
	zenaChainID := "1234"

	bufferedCheckpoint := hmTypes.CreateBlock(
		startBlock,
		endBlock,
		rootHash,
		proposerAddress,
		zenaChainID,
		timestamp,
	)

	Checkpoints := make([]hmTypes.Checkpoint, ackCount)

	for i := range Checkpoints {
		Checkpoints[i] = bufferedCheckpoint
	}

	params := types.DefaultParams()
	genesisState := types.NewGenesisState(
		params,
		&bufferedCheckpoint,
		uint64(lastNoACK),
		uint64(ackCount),
		Checkpoints,
	)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesisState)
}
