package types

import (
	"fmt"
)

// DividendAccountProof contains ID, merkle proof, leaf index in merketree
type DividendAccountProof struct {
	User  IrisAddress `json:"user"`
	Proof HexBytes        `json:"accountProof"`
	Index uint64          `json:"index"`
}

// NewDividendAccountProof generate proof for new dividend account
func NewDividendAccountProof(user IrisAddress, proof HexBytes, index uint64) DividendAccountProof {
	return DividendAccountProof{
		User:  user,
		Proof: proof,
		Index: index,
	}
}

func (ap *DividendAccountProof) String() string {
	if ap == nil {
		return "nil-DividendAccountProof"
	}

	return fmt.Sprintf(
		"DividendAccount{%v %v %v}",
		ap.User,
		ap.Proof,
		ap.Index,
	)
}
