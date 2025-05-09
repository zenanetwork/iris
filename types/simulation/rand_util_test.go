package simulation_test

import (
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/zenanetwork/iris/types/simulation"
)

func TestRandSubsetCoins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		r     *rand.Rand
		coins sdk.Coins
	}{
		{"seed=1", rand.New(rand.NewSource(1)), mustParseCoins("100stake,2testtoken")},
		{"seed=50", rand.New(rand.NewSource(50)), mustParseCoins("100stake,2testtoken")},
		{"seed=99", rand.New(rand.NewSource(99)), mustParseCoins("100stake,2testtoken")},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := simulation.RandSubsetCoins(tt.r, tt.coins)
			gotStringRep := got.String()
			sortedStringRep := got.Sort().String()
			require.Equal(t, gotStringRep, sortedStringRep)
		})
	}
}

func TestRandStringOfLength(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewSource(time.Now().Unix()))
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"0-size", 0, 0},
		{"10-size", 10, 10},
		{"10-size", 1_000_000_000, 1_000_000_000},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := simulation.RandStringOfLength(r, tt.n)
			require.Equal(t, tt.want, len(got))
		})
	}
}

func mustParseCoins(s string) sdk.Coins {
	coins, err := sdk.ParseCoins(s)
	if err != nil {
		panic(err)
	}

	return coins
}
