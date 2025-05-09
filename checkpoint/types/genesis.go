package types

import (
	"encoding/json"
	"fmt"
	"math"

	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/zena/types"
)

// GenesisState is the checkpoint state that must be provided at genesis.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`

	BufferedCheckpoint *hmTypes.Checkpoint  `json:"buffered_checkpoint" yaml:"buffered_checkpoint"`
	LastNoACK          uint64               `json:"last_no_ack" yaml:"last_no_ack"`
	AckCount           uint64               `json:"ack_count" yaml:"ack_count"`
	Checkpoints        []hmTypes.Checkpoint `json:"checkpoints" yaml:"checkpoints"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(
	params Params,
	bufferedCheckpoint *hmTypes.Checkpoint,
	lastNoACK uint64,
	ackCount uint64,
	checkpoints []hmTypes.Checkpoint,
) GenesisState {
	return GenesisState{
		Params:             params,
		BufferedCheckpoint: bufferedCheckpoint,
		LastNoACK:          lastNoACK,
		AckCount:           ackCount,
		Checkpoints:        checkpoints,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis performs basic validation of zena genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	if data.AckCount > math.MaxInt {
		return fmt.Errorf("ack count value out of range for int: %d", data.AckCount)
	}
	if int(data.AckCount) != len(data.Checkpoints) {
		return fmt.Errorf("ack count does not match the number of checkpoints")
	}

	return nil
}

// GetGenesisStateFromAppState returns staking GenesisState given raw application genesis state
func GetGenesisStateFromAppState(appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		types.ModuleCdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}
