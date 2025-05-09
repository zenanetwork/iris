//nolint:gosec
package simulation

import (
	"crypto/rand"
	"math/big"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/zenanetwork/iris/bridge/setu/util"
	"github.com/zenanetwork/iris/types"
)

// GenRandomVal generate random validators
func GenRandomVal(count int, startBlock uint64, power int64, timeAlive uint64, randomise bool, startID uint64, nonce uint64) (validators []types.Validator) {
	for i := 0; i < count; i++ {
		privKey1 := secp256k1.GenPrivKey()
		pubkey := types.NewPubKey(util.AppendPrefix(privKey1.PubKey().Bytes()))

		if randomise {
			startBlock = generateRandNumber(10)
			power = int64(generateRandNumber(100))
		}

		newVal := types.Validator{
			ID:               types.NewValidatorID(startID + uint64(i)),
			StartEpoch:       startBlock,
			EndEpoch:         startBlock + timeAlive,
			VotingPower:      power,
			Signer:           types.HexToIrisAddress(pubkey.Address().String()),
			PubKey:           pubkey,
			ProposerPriority: 0,
			Nonce:            nonce,
		}
		validators = append(validators, newVal)
	}

	return
}

func generateRandNumber(maxV int64) uint64 {
	nBig, err := rand.Int(rand.Reader, big.NewInt(maxV))
	if err != nil {
		return 1
	}

	return nBig.Uint64()
}
