package types

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zenanetwork/iris/helper"
	"github.com/zenanetwork/iris/params/subspace"
	hmTypes "github.com/zenanetwork/iris/types"
)

// Default parameter values
const (
	DefaultMainchainTxConfirmations           uint64 = 6
	DefaultMaticchainTxConfirmations          uint64 = 10
	DefaultMaticchainMilestoneTxConfirmations uint64 = 16
)

var (
	DefaultStateReceiverAddress hmTypes.IrisAddress = hmTypes.HexToIrisAddress("0x0000000000000000000000000000000000001001")
	DefaultValidatorSetAddress  hmTypes.IrisAddress = hmTypes.HexToIrisAddress("0x0000000000000000000000000000000000001000")
)

// Parameter keys
var (
	KeyMainchainTxConfirmations  = []byte("MainchainTxConfirmations")
	KeyMaticchainTxConfirmations = []byte("MaticchainTxConfirmations")
	KeyChainParams               = []byte("ChainParams")
)

var _ subspace.ParamSet = &Params{}

// ChainParams chain related params
type ChainParams struct {
	ZenaChainID           string              `json:"zena_chain_id" yaml:"zena_chain_id"`
	MaticTokenAddress     hmTypes.IrisAddress `json:"matic_token_address" yaml:"matic_token_address"`
	StakingManagerAddress hmTypes.IrisAddress `json:"staking_manager_address" yaml:"staking_manager_address"`
	SlashManagerAddress   hmTypes.IrisAddress `json:"slash_manager_address" yaml:"slash_manager_address"`
	RootChainAddress      hmTypes.IrisAddress `json:"root_chain_address" yaml:"root_chain_address"`
	StakingInfoAddress    hmTypes.IrisAddress `json:"staking_info_address" yaml:"staking_info_address"`
	StateSenderAddress    hmTypes.IrisAddress `json:"state_sender_address" yaml:"state_sender_address"`

	// Zena Chain Contracts
	StateReceiverAddress hmTypes.IrisAddress `json:"state_receiver_address" yaml:"state_receiver_address"`
	ValidatorSetAddress  hmTypes.IrisAddress `json:"validator_set_address" yaml:"validator_set_address"`
}

func (cp ChainParams) String() string {
	return fmt.Sprintf(`
	ZenaChainID: 									%s
  MaticTokenAddress:            %s
	StakingManagerAddress:        %s
	SlashManagerAddress:        %s
	RootChainAddress:             %s
  StakingInfoAddress:           %s
	StateSenderAddress:           %s
	StateReceiverAddress: 				%s
	ValidatorSetAddress:					%s`,
		cp.ZenaChainID, cp.MaticTokenAddress, cp.StakingManagerAddress, cp.SlashManagerAddress, cp.RootChainAddress, cp.StakingInfoAddress, cp.StateSenderAddress, cp.StateReceiverAddress, cp.ValidatorSetAddress)
}

// Params defines the parameters for the chainmanager module.
type Params struct {
	MainchainTxConfirmations  uint64      `json:"mainchain_tx_confirmations" yaml:"mainchain_tx_confirmations"`
	MaticchainTxConfirmations uint64      `json:"maticchain_tx_confirmations" yaml:"maticchain_tx_confirmations"`
	ChainParams               ChainParams `json:"chain_params" yaml:"chain_params"`
}

// NewParams creates a new Params object
func NewParams(mainchainTxConfirmations uint64, maticchainTxConfirmations uint64, chainParams ChainParams) Params {
	return Params{
		MainchainTxConfirmations:  mainchainTxConfirmations,
		MaticchainTxConfirmations: maticchainTxConfirmations,
		ChainParams:               chainParams,
	}
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyMainchainTxConfirmations, &p.MainchainTxConfirmations},
		{KeyMaticchainTxConfirmations, &p.MaticchainTxConfirmations},
		{KeyChainParams, &p.ChainParams},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)

	return bytes.Equal(bz1, bz2)
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder

	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("MainchainTxConfirmations: %d\n", p.MainchainTxConfirmations))
	sb.WriteString(fmt.Sprintf("MaticchainTxConfirmations: %d\n", p.MaticchainTxConfirmations))
	sb.WriteString(fmt.Sprintf("ChainParams: %s\n", p.ChainParams.String()))

	return sb.String()
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateIrisAddress("matic_token_address", p.ChainParams.MaticTokenAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("staking_manager_address", p.ChainParams.StakingManagerAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("slash_manager_address", p.ChainParams.SlashManagerAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("root_chain_address", p.ChainParams.RootChainAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("staking_info_address", p.ChainParams.StakingInfoAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("state_sender_address", p.ChainParams.StateSenderAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("state_receiver_address", p.ChainParams.StateReceiverAddress); err != nil {
		return err
	}

	if err := validateIrisAddress("validator_set_address", p.ChainParams.ValidatorSetAddress); err != nil {
		return err
	}

	return nil
}

func validateIrisAddress(key string, value hmTypes.IrisAddress) error {
	if value.String() == "" {
		return fmt.Errorf("Invalid value %s in chain_params", key)
	}

	return nil
}

//
// Extra functions
//

// ParamKeyTable for auth module
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MainchainTxConfirmations:  DefaultMainchainTxConfirmations,
		MaticchainTxConfirmations: DefaultMaticchainTxConfirmations,
		ChainParams: ChainParams{
			ZenaChainID:          helper.DefaultZenaChainID,
			StateReceiverAddress: DefaultStateReceiverAddress,
			ValidatorSetAddress:  DefaultValidatorSetAddress,
		},
	}
}
