package types

import (
	"bytes"
	"encoding/hex"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"

	"github.com/zenanetwork/go-zenanet/common"
)

// Ensure that different address types implement the interface
var _ yaml.Marshaler = IrisHash{}

// IrisHash represents iris address
type IrisHash common.Hash

// ZeroIrisHash represents zero address
var ZeroIrisHash = IrisHash{}

// EthHash get eth hash
func (aa IrisHash) EthHash() common.Hash {
	return common.Hash(aa)
}

// Equals returns boolean for whether two IrisHash are Equal
func (aa IrisHash) Equals(aa2 IrisHash) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Empty returns boolean for whether an AccAddress is empty
func (aa IrisHash) Empty() bool {
	return bytes.Equal(aa.Bytes(), ZeroIrisHash.Bytes())
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa IrisHash) Marshal() ([]byte, error) {
	return aa.Bytes(), nil
}

// Unmarshal sets the address to the given data. It is needed for protobuf
// compatibility.
func (aa *IrisHash) Unmarshal(data []byte) error {
	*aa = IrisHash(common.BytesToHash(data))
	return nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa IrisHash) MarshalJSON() ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(aa.String())
}

// MarshalYAML marshals to YAML using Bech32.
func (aa IrisHash) MarshalYAML() (interface{}, error) {
	return aa.String(), nil
}

// UnmarshalJSON unmarshals from JSON assuming Bech32 encoding.
func (aa *IrisHash) UnmarshalJSON(data []byte) error {
	var s string
	if err := jsoniter.ConfigFastest.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToIrisHash(s)

	return nil
}

// UnmarshalYAML unmarshals from JSON assuming Bech32 encoding.
func (aa *IrisHash) UnmarshalYAML(data []byte) error {
	var s string
	if err := yaml.Unmarshal(data, &s); err != nil {
		return err
	}

	*aa = HexToIrisHash(s)

	return nil
}

// Bytes returns the raw address bytes.
func (aa IrisHash) Bytes() []byte {
	return aa[:]
}

// String implements the Stringer interface.
func (aa IrisHash) String() string {
	if aa.Empty() {
		return ""
	}

	return "0x" + hex.EncodeToString(aa.Bytes())
}

// Hex returns hex string
func (aa IrisHash) Hex() string {
	return aa.String()
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (aa IrisHash) Format(s fmt.State, verb rune) {
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
// hash utils
//

// BytesToIrisHash returns Address with value b.
func BytesToIrisHash(b []byte) IrisHash {
	return IrisHash(common.BytesToHash(b))
}

// HexToIrisHash returns Address with value b.
func HexToIrisHash(b string) IrisHash {
	return IrisHash(common.HexToHash(b))
}
