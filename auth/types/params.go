package types

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/zenanetwork/iris/params/subspace"
)

// Default parameter values
const (
	DefaultMaxMemoCharacters      uint64 = 256
	DefaultTxSigLimit             uint64 = 7
	DefaultTxSizeCostPerByte      uint64 = 10
	DefaultSigVerifyCostED25519   uint64 = 590
	DefaultSigVerifyCostSecp256k1 uint64 = 1000

	DefaultMaxTxGas uint64 = 1000000
	DefaultTxFees   string = "1000000000000000"
)

// Parameter keys
var (
	KeyMaxMemoCharacters      = []byte("MaxMemoCharacters")
	KeyTxSigLimit             = []byte("TxSigLimit")
	KeyTxSizeCostPerByte      = []byte("TxSizeCostPerByte")
	KeySigVerifyCostED25519   = []byte("SigVerifyCostED25519")
	KeySigVerifyCostSecp256k1 = []byte("SigVerifyCostSecp256k1")

	KeyMaxTxGas = []byte("MaxTxGas")
	KeyTxFees   = []byte("TxFees")
)

var _ subspace.ParamSet = &Params{}

// Params defines the parameters for the auth module.
type Params struct {
	MaxMemoCharacters      uint64 `json:"max_memo_characters" yaml:"max_memo_characters"`
	TxSigLimit             uint64 `json:"tx_sig_limit" yaml:"tx_sig_limit"`
	TxSizeCostPerByte      uint64 `json:"tx_size_cost_per_byte" yaml:"tx_size_cost_per_byte"`
	SigVerifyCostED25519   uint64 `json:"sig_verify_cost_ed25519" yaml:"sig_verify_cost_ed25519"`
	SigVerifyCostSecp256k1 uint64 `json:"sig_verify_cost_secp256k1" yaml:"sig_verify_cost_secp256k1"`

	MaxTxGas uint64 `json:"max_tx_gas" yaml:"max_tx_gas"`
	TxFees   string `json:"tx_fees" yaml:"tx_fees"`
}

// NewParams creates a new Params object
func NewParams(
	maxMemoCharacters uint64,
	txSigLimit uint64,
	txSizeCostPerByte uint64,
	sigVerifyCostED25519 uint64,
	sigVerifyCostSecp256k1 uint64,

	maxTxGas uint64,
	txFees string,
) Params {
	return Params{
		MaxMemoCharacters:      maxMemoCharacters,
		TxSigLimit:             txSigLimit,
		TxSizeCostPerByte:      txSizeCostPerByte,
		SigVerifyCostED25519:   sigVerifyCostED25519,
		SigVerifyCostSecp256k1: sigVerifyCostSecp256k1,

		MaxTxGas: maxTxGas,
		TxFees:   txFees,
	}
}

// ParamKeyTable for auth module
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyMaxMemoCharacters, &p.MaxMemoCharacters},
		{KeyTxSigLimit, &p.TxSigLimit},
		{KeyTxSizeCostPerByte, &p.TxSizeCostPerByte},
		{KeySigVerifyCostED25519, &p.SigVerifyCostED25519},
		{KeySigVerifyCostSecp256k1, &p.SigVerifyCostSecp256k1},

		{KeyMaxTxGas, &p.MaxTxGas},
		{KeyTxFees, &p.TxFees},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)

	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxMemoCharacters:      DefaultMaxMemoCharacters,
		TxSigLimit:             DefaultTxSigLimit,
		TxSizeCostPerByte:      DefaultTxSizeCostPerByte,
		SigVerifyCostED25519:   DefaultSigVerifyCostED25519,
		SigVerifyCostSecp256k1: DefaultSigVerifyCostSecp256k1,

		MaxTxGas: DefaultMaxTxGas,
		TxFees:   DefaultTxFees,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder

	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("MaxMemoCharacters: %d\n", p.MaxMemoCharacters))
	sb.WriteString(fmt.Sprintf("TxSigLimit: %d\n", p.TxSigLimit))
	sb.WriteString(fmt.Sprintf("TxSizeCostPerByte: %d\n", p.TxSizeCostPerByte))
	sb.WriteString(fmt.Sprintf("SigVerifyCostED25519: %d\n", p.SigVerifyCostED25519))
	sb.WriteString(fmt.Sprintf("SigVerifyCostSecp256k1: %d\n", p.SigVerifyCostSecp256k1))
	sb.WriteString(fmt.Sprintf("MaxTxGas: %d\n", p.MaxTxGas))
	sb.WriteString(fmt.Sprintf("TxFees: %s\n", p.TxFees))

	return sb.String()
}

func validateTxSigLimit(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid tx signature limit: %d", v)
	}

	return nil
}

func validateSigVerifyCostED25519(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid ED25519 signature verification cost: %d", v)
	}

	return nil
}

func validateSigVerifyCostSecp256k1(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid SECK256k1 signature verification cost: %d", v)
	}

	return nil
}

func validateTxSizeCostPerByte(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid tx size cost per byte: %d", v)
	}

	return nil
}

func validateMaxTxGas(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("invalid max tx gas: %d", v)
	}

	return nil
}

func validateTxFees(v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("invalid tx fees: %s", v)
	}

	if _, ok := big.NewInt(0).SetString(v, 10); !ok {
		return fmt.Errorf("invalid tx fees: %s, should be valid big integer", v)
	}

	return nil
}

// Validate checks that the parameters have valid values.
func (p Params) Validate() error {
	if err := validateTxSigLimit(p.TxSigLimit); err != nil {
		return err
	}

	if err := validateSigVerifyCostED25519(p.SigVerifyCostED25519); err != nil {
		return err
	}

	if err := validateSigVerifyCostSecp256k1(p.SigVerifyCostSecp256k1); err != nil {
		return err
	}

	if err := validateSigVerifyCostSecp256k1(p.MaxMemoCharacters); err != nil {
		return err
	}

	if err := validateTxSizeCostPerByte(p.TxSizeCostPerByte); err != nil {
		return err
	}

	if err := validateMaxTxGas(p.MaxTxGas); err != nil {
		return err
	}

	if err := validateTxFees(p.TxFees); err != nil {
		return err
	}

	return nil
}
