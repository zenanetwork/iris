//nolint:gosec
package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/zenanetwork/iris/chainmanager/types"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/types/module"
	"github.com/zenanetwork/iris/types/simulation"
)

// Parameter keys
const (
	MainchainTxConfirmations  = "mainchain_tx_confirmations"
	MaticchainTxConfirmations = "maticchain_tx_confirmations"

	ZenaChainID           = "zena_chain_id"
	MaticTokenAddress     = "matic_token_address"     //nolint
	StakingManagerAddress = "staking_manager_address" //nolint
	SlashManagerAddress   = "slash_manager_address"   //nolint
	RootChainAddress      = "root_chain_address"      //nolint
	StakingInfoAddress    = "staking_info_address"    //nolint
	StateSenderAddress    = "state_sender_address"    //nolint

	// Zena Chain Contracts
	StateReceiverAddress = "state_receiver_address" //nolint
	ValidatorSetAddress  = "validator_set_address"  //nolint
)

func GenMainchainTxConfirmations(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 100))
}

func GenMaticchainTxConfirmations(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 100))
}

func GenIrisAddress() hmTypes.IrisAddress {
	return hmTypes.BytesToIrisAddress(simulation.RandHex(20))
}

// GenZenaChainId returns randomc chainID
func GenZenaChainId(r *rand.Rand) string {
	return strconv.Itoa(simulation.RandIntBetween(r, 0, math.MaxInt32))
}

func RandomizedGenState(simState *module.SimulationState) {
	var mainchainTxConfirmations uint64

	simState.AppParams.GetOrGenerate(simState.Cdc, MainchainTxConfirmations, &mainchainTxConfirmations, simState.Rand,
		func(r *rand.Rand) { mainchainTxConfirmations = GenMainchainTxConfirmations(r) },
	)

	var (
		maticchainTxConfirmations uint64
		zenaChainID               string
	)

	simState.AppParams.GetOrGenerate(simState.Cdc, MaticchainTxConfirmations, &maticchainTxConfirmations, simState.Rand,
		func(r *rand.Rand) { maticchainTxConfirmations = GenMaticchainTxConfirmations(r) },
	)

	simState.AppParams.GetOrGenerate(simState.Cdc, ZenaChainID, &zenaChainID, simState.Rand,
		func(r *rand.Rand) { zenaChainID = GenZenaChainId(r) },
	)

	var (
		maticTokenAddress     = GenIrisAddress()
		stakingManagerAddress = GenIrisAddress()
		slashManagerAddress   = GenIrisAddress()
		rootChainAddress      = GenIrisAddress()
		stakingInfoAddress    = GenIrisAddress()
		stateSenderAddress    = GenIrisAddress()
		stateReceiverAddress  = GenIrisAddress()
		validatorSetAddress   = GenIrisAddress()
	)

	chainParams := types.ChainParams{
		ZenaChainID:           zenaChainID,
		MaticTokenAddress:     maticTokenAddress,
		StakingManagerAddress: stakingManagerAddress,
		SlashManagerAddress:   slashManagerAddress,
		RootChainAddress:      rootChainAddress,
		StakingInfoAddress:    stakingInfoAddress,
		StateSenderAddress:    stateSenderAddress,
		StateReceiverAddress:  stateReceiverAddress,
		ValidatorSetAddress:   validatorSetAddress,
	}
	params := types.NewParams(mainchainTxConfirmations, maticchainTxConfirmations, chainParams)
	chainManagerGenesis := types.NewGenesisState(params)
	fmt.Printf("Selected randomly generated chainmanager parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, chainManagerGenesis))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(chainManagerGenesis)
}
