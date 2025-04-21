package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// AddrLen defines a valid address length
	AddrLen = 20
)

// Ensure that different address types implement the interface
var _ sdk.Address = IrisAddress{}
var _ yaml.Marshaler = IrisAddress{}

// IrisAddress represents iris address
type IrisAddress common.Address

// ZeroIrisAddress represents zero address
var ZeroIrisAddress = IrisAddress{}

// EthAddress get eth address
func (aa IrisAddress) EthAddress() common.Address {
	return common.Address(aa)
}

// Equals returns boolean for whether two AccAddresses are Equal
func (aa IrisAddress) Equals(aa2 sdk.Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Empty returns boolean for whether an AccAddress is empty
func (aa IrisAddress) Empty() bool {
	return bytes.Equal(aa.Bytes(), ZeroIrisAddress.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa IrisAddress) Marshal() ([]byte, error) {
	return aa.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *IrisAddress) Unmarshal(data []byte) error {
	*aa = IrisAddress(common.BytesToAddress(data))
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa IrisAddress) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(aa.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (aa IrisAddress) MarshalYAML() (interface{}, error) {
	return aa.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *IrisAddress) UnmarshalJSON(data []byte) error {
	var s string
	if err := jsoniter.ConfigFastest.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToIrisAddress(s)

	return nil
}

// UnmarshalYAML unmarshals from JSON assuming Bech32 encoding.
func (aa *IrisAddress) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToIrisAddress(s)

	return nil
}

// Bytes returns the raw address bytes.
func (aa IrisAddress) Bytes() []byte {
	return aa[:]
}

// String implements the Stringer interface.
func (aa IrisAddress) String() string {
	return "0x" + hex.EncodeToString(aa.Bytes())
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (aa IrisAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(aa.String()))
	case 'p':
		s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		s.Write([]byte(fmt.Sprintf("%X", aa.Bytes())))
	}
}

//
// Address utils
//

// BytesToIrisAddress returns Address with value b.
func BytesToIrisAddress(b []byte) IrisAddress {
	return IrisAddress(common.BytesToAddress(b))
}

// HexToIrisAddress returns Address with value b.
func HexToIrisAddress(b string) IrisAddress {
	return IrisAddress(common.HexToAddress(b))
}

// AccAddressToIrisAddress returns Address with value b.
func AccAddressToIrisAddress(b sdk.AccAddress) IrisAddress {
	return BytesToIrisAddress(b[:])
}

// IrisAddressToAccAddress returns Address with value b.
func IrisAddressToAccAddress(b IrisAddress) sdk.AccAddress {
	return sdk.AccAddress(b.Bytes())
}

// SampleIrisAddress returns sample address
func SampleIrisAddress(s string) IrisAddress {
	return BytesToIrisAddress([]byte(s))
}
