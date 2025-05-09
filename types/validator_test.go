package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// valInput struct is used to seed data for testing
// if the need arises it can be ported to the main build
type valInput struct {
	id         ValidatorID
	startEpoch uint64
	endEpoch   uint64
	power      int64
	nonce      uint64
	pubKey     PubKey
	signer     IrisAddress
}

func TestNewValidator(t *testing.T) {
	t.Parallel()

	// valCase created so as to pass it to assertPanics func,
	// ideally would like to get rid of this and pass the function directly
	tc := []struct {
		in  valInput
		out *Validator
		msg string
	}{
		{
			in: valInput{
				id:     ValidatorID(uint64(0)),
				signer: BytesToIrisAddress([]byte("12345678909876543210")),
				nonce:  uint64(0),
			},
			out: &Validator{Signer: BytesToIrisAddress([]byte("12345678909876543210")), Nonce: uint64(0)},
			msg: "testing for exact IrisAddress",
		},
		{
			in: valInput{
				id:     ValidatorID(uint64(0)),
				signer: BytesToIrisAddress([]byte("1")),
				nonce:  uint64(1),
			},
			out: &Validator{Signer: BytesToIrisAddress([]byte("1")), Nonce: uint64(1)},
			msg: "testing for small IrisAddress",
		},
		{
			in: valInput{
				id:     ValidatorID(uint64(0)),
				signer: BytesToIrisAddress([]byte("123456789098765432101")),
				nonce:  uint64(32),
			},
			out: &Validator{Signer: BytesToIrisAddress([]byte("123456789098765432101")), Nonce: uint64(32)},
			msg: "testing for excessively long IrisAddress, max length is supposed to be 20",
		},
	}
	for _, c := range tc {
		out := NewValidator(c.in.id, c.in.startEpoch, c.in.endEpoch, c.in.nonce, c.in.power, c.in.pubKey, c.in.signer)
		assert.Equal(t, c.out, out)
	}
}

// TestSortValidatorByAddress am populating only the signer as that is the only value used in sorting
func TestSortValidatorByAddress(t *testing.T) {
	t.Parallel()

	tc := []struct {
		in  []Validator
		out []Validator
		msg string
	}{
		{
			in: []Validator{
				{Signer: BytesToIrisAddress([]byte("3"))},
				{Signer: BytesToIrisAddress([]byte("2"))},
				{Signer: BytesToIrisAddress([]byte("1"))},
			},
			out: []Validator{
				{Signer: BytesToIrisAddress([]byte("1"))},
				{Signer: BytesToIrisAddress([]byte("2"))},
				{Signer: BytesToIrisAddress([]byte("3"))},
			},
			msg: "reverse sorting of validator objects",
		},
	}
	for i, c := range tc {
		out := SortValidatorByAddress(c.in)
		assert.Equal(t, c.out, out, fmt.Sprintf("i: %v, case: %v", i, c.msg))
	}
}

func TestValidateBasic(t *testing.T) {
	t.Parallel()

	tc := []struct {
		in  Validator
		out bool
		msg string
	}{
		{
			in:  Validator{StartEpoch: 1, EndEpoch: 5, Nonce: 0, PubKey: NewPubKey([]byte("nonZeroTestPubKey")), Signer: BytesToIrisAddress([]byte("3"))},
			out: true,
			msg: "Valid basic validator test",
		},
		{
			in:  Validator{StartEpoch: 1, EndEpoch: 5, Nonce: 0, PubKey: NewPubKey([]byte("")), Signer: BytesToIrisAddress([]byte("3"))},
			out: false,
			msg: "Invalid PubKey \"\"",
		},
		{
			in:  Validator{StartEpoch: 1, EndEpoch: 5, Nonce: 0, PubKey: ZeroPubKey, Signer: BytesToIrisAddress([]byte("3"))},
			out: false,
			msg: "Invalid PubKey",
		},
		{
			in:  Validator{StartEpoch: 1, EndEpoch: 1, Nonce: 0, PubKey: NewPubKey([]byte("nonZeroTestPubKey")), Signer: BytesToIrisAddress([]byte(""))},
			out: false,
			msg: "Invalid Signer",
		},
		{
			in:  Validator{},
			out: false,
			msg: "Invalid basic validator test",
		},
	}

	for _, c := range tc {
		out := c.in.ValidateBasic()
		assert.Equal(t, c.out, out, c.msg)
	}
}
