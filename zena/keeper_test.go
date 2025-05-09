package zena_test

import (
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/zenanetwork/go-zenanet/common"
	ethTypes "github.com/zenanetwork/go-zenanet/core/types"
	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/helper/mocks"
	hmTypes "github.com/zenanetwork/iris/types"
	"github.com/zenanetwork/iris/zena"
)

type ZenaKeeperTestSuite struct {
	suite.Suite

	app            *app.IrisApp
	contractCaller *mocks.IContractCaller
	ctx            sdk.Context
}

func createTestApp(isCheckTx bool) (*app.IrisApp, sdk.Context) {
	app := app.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})

	return app, ctx
}

func (suite *ZenaKeeperTestSuite) SetupTest() {
	suite.app, suite.ctx = createTestApp(false)
	suite.contractCaller = &mocks.IContractCaller{}
	suite.app.ZenaKeeper.SetContractCaller(suite.contractCaller)
}

func TestKeeperTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ZenaKeeperTestSuite))
}

func (s *ZenaKeeperTestSuite) TestGetNextSpanSeed() {
	require, ctx, zenaKeeper := s.Require(), s.ctx, s.app.ZenaKeeper
	valSet := s.setupValSet()
	vals := make([]hmTypes.Validator, 0, len(valSet.Validators))
	for _, val := range valSet.Validators {
		vals = append(vals, *val)
	}

	spans := []hmTypes.Span{
		hmTypes.NewSpan(0, 0, 256, *valSet, vals, "test-chain"),
		hmTypes.NewSpan(1, 257, 6656, *valSet, vals, "test-chain"),
		hmTypes.NewSpan(2, 6657, 16656, *valSet, vals, "test-chain"),
		hmTypes.NewSpan(3, 16657, 26656, *valSet, vals, "test-chain"),
	}

	for _, span := range spans {
		err := zenaKeeper.AddNewSpan(ctx, span)
		require.NoError(err)
	}

	val1Addr := vals[0].PubKey.Address()
	val2Addr := vals[1].PubKey.Address()
	val3Addr := vals[2].PubKey.Address()

	zenaParams := zenaKeeper.GetParams(ctx)

	seedBlock1 := spans[0].EndBlock
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(seedBlock1))).Return(&val2Addr, nil)

	seedBlock2 := spans[1].EndBlock - zenaParams.SprintDuration
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(spans[1].EndBlock))).Return(&val2Addr, nil)
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(seedBlock2))).Return(&val1Addr, nil)
	for block := spans[1].EndBlock - (2 * zenaParams.SprintDuration); block >= spans[1].StartBlock; block -= zenaParams.SprintDuration {
		s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(block))).Return(&val1Addr, nil)
	}

	seedBlock3 := spans[2].EndBlock - (2 * zenaParams.SprintDuration)
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(spans[2].EndBlock))).Return(&val1Addr, nil)
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(spans[2].EndBlock-zenaParams.SprintDuration))).Return(&val2Addr, nil)
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(seedBlock3))).Return(&val3Addr, nil)

	seedBlock4 := spans[3].EndBlock - zenaParams.SprintDuration
	s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(spans[3].EndBlock))).Return(&val1Addr, nil)

	for block := spans[3].EndBlock; block >= spans[3].StartBlock; block -= zenaParams.SprintDuration {
		s.contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(int64(block))).Return(&val2Addr, nil)
	}

	blockHeader1 := ethTypes.Header{Number: big.NewInt(int64(seedBlock1))}
	blockHash1 := blockHeader1.Hash()
	blockHeader2 := ethTypes.Header{Number: big.NewInt(int64(seedBlock2))}
	blockHash2 := blockHeader2.Hash()
	blockHeader3 := ethTypes.Header{Number: big.NewInt(int64(seedBlock3))}
	blockHash3 := blockHeader3.Hash()
	blockHeader4 := ethTypes.Header{Number: big.NewInt(int64(seedBlock4))}
	blockHash4 := blockHeader4.Hash()

	s.contractCaller.On("GetMaticChainBlock", big.NewInt(int64(seedBlock1))).Return(&blockHeader1, nil)
	s.contractCaller.On("GetMaticChainBlock", big.NewInt(int64(seedBlock2))).Return(&blockHeader2, nil)
	s.contractCaller.On("GetMaticChainBlock", big.NewInt(int64(seedBlock3))).Return(&blockHeader3, nil)
	s.contractCaller.On("GetMaticChainBlock", big.NewInt(int64(seedBlock4))).Return(&blockHeader4, nil)

	testcases := []struct {
		name             string
		lastSpanId       uint64
		lastSeedProducer *common.Address
		expSeed          common.Hash
		expAuthor        *common.Address
	}{
		{
			name:             "Last seed producer is different than end block author",
			lastSeedProducer: &val2Addr,
			lastSpanId:       0,
			expSeed:          blockHash1,
			expAuthor:        &val2Addr,
		},
		{
			name:             "Last seed producer is same as end block author",
			lastSeedProducer: &val1Addr,
			lastSpanId:       1,
			expSeed:          blockHash2,
			expAuthor:        &val1Addr,
		},
		{
			name:             "Next seed producer should be different from previous recent seed producers",
			lastSeedProducer: &val2Addr,
			lastSpanId:       2,
			expSeed:          blockHash3,
			expAuthor:        &val3Addr,
		},
		{
			name:             "If no unique seed producer is found, first block with different author from previous seed producer is selected",
			lastSeedProducer: &val1Addr,
			lastSpanId:       3,
			expSeed:          blockHash4,
			expAuthor:        &val2Addr,
		},
	}

	lastSpanID := uint64(0)

	for _, tc := range testcases {
		err := zenaKeeper.StoreSeedProducer(ctx, tc.lastSpanId, tc.lastSeedProducer)
		require.NoError(err)

		lastSpanID = tc.lastSpanId
	}

	err := zenaKeeper.StoreSeedProducer(ctx, lastSpanID+1, &val1Addr)
	require.NoError(err)

	for _, tc := range testcases {
		s.T().Run(tc.name, func(t *testing.T) {
			seed, author, err := zenaKeeper.GetNextSpanSeed(ctx, tc.lastSpanId+2)
			require.NoError(err)
			require.Equal(tc.expSeed.Bytes(), seed.Bytes())
			require.Equal(tc.expAuthor.Bytes(), author.Bytes())
		})
	}
}

func (s *ZenaKeeperTestSuite) TestProposeSpanOne() {
	app, ctx := createTestApp(false)
	contractCaller := &mocks.IContractCaller{}
	app.ZenaKeeper.SetContractCaller(contractCaller)

	valSet := setupValSet()
	vals := make([]hmTypes.Validator, 0, len(valSet.Validators))
	for _, val := range valSet.Validators {
		vals = append(vals, *val)
	}
	err := app.ZenaKeeper.AddNewSpan(ctx, hmTypes.NewSpan(0, 0, 256, *valSet, vals, "test-chain"))
	s.Require().NoError(err)

	val1Addr := vals[0].PubKey.Address()

	seedBlock1 := int64(1)
	contractCaller.On("GetZenaChainBlockAuthor", big.NewInt(seedBlock1)).Return(&val1Addr, nil)

	blockHeader1 := ethTypes.Header{Number: big.NewInt(seedBlock1)}
	blockHash1 := blockHeader1.Hash()
	contractCaller.On("GetMaticChainBlock", big.NewInt(seedBlock1)).Return(&blockHeader1, nil)

	seed, author, err := app.ZenaKeeper.GetNextSpanSeed(ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(blockHash1.Bytes(), seed.Bytes())
	s.Require().Equal(val1Addr.Bytes(), author.Bytes())
}

func (s *ZenaKeeperTestSuite) TestGetSeedProducer() {
	zenaKeeper := s.app.ZenaKeeper
	producer := common.HexToAddress("0xc0ffee254729296a45a3885639AC7E10F9d54979")
	err := zenaKeeper.StoreSeedProducer(s.ctx, 1, &producer)
	s.Require().NoError(err)

	author, err := zenaKeeper.GetSeedProducer(s.ctx, 1)
	s.Require().NoError(err)
	s.Require().Equal(producer.Bytes(), author.Bytes())

}

func (s *ZenaKeeperTestSuite) TestRollbackVotingPowers() {
	testcases := []struct {
		name    string
		valsOld []hmTypes.Validator
		valsNew []hmTypes.Validator
		expRes  []hmTypes.Validator
	}{
		{
			name:    "Revert to old voting powers",
			valsOld: []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}, {ID: 3, VotingPower: 300}},
			valsNew: []hmTypes.Validator{{ID: 1, VotingPower: 200}, {ID: 2, VotingPower: 400}, {ID: 3, VotingPower: 200}},
			expRes:  []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}, {ID: 3, VotingPower: 300}},
		},
		{
			name:    "Validator joined",
			valsOld: []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}},
			valsNew: []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}, {ID: 3, VotingPower: 300}},
			expRes:  []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}, {ID: 3, VotingPower: 0}},
		},
		{
			name:    "Validator left",
			valsOld: []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}, {ID: 3, VotingPower: 300}},
			valsNew: []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}},
			expRes:  []hmTypes.Validator{{ID: 1, VotingPower: 100}, {ID: 2, VotingPower: 200}},
		},
	}

	for _, tc := range testcases {
		s.T().Run(tc.name, func(t *testing.T) {
			res := zena.RollbackVotingPowers(s.ctx, tc.valsNew, tc.valsOld)
			s.Require().Equal(tc.expRes, res)
		})
	}
}

func (suite *ZenaKeeperTestSuite) setupValSet() *hmTypes.ValidatorSet {
	suite.T().Helper()
	return setupValSet()
}

func setupValSet() *hmTypes.ValidatorSet {

	pubKey1 := hmTypes.NewPubKey([]byte("pubkey1"))
	pubKey2 := hmTypes.NewPubKey([]byte("pubkey2"))
	pubKey3 := hmTypes.NewPubKey([]byte("pubkey3"))

	val1 := hmTypes.NewValidator(1, 100, 0, 1, 100, hmTypes.NewPubKey(pubKey1[:]), hmTypes.IrisAddress(pubKey1.Address()))
	val2 := hmTypes.NewValidator(2, 100, 0, 1, 100, hmTypes.NewPubKey(pubKey2[:]), hmTypes.IrisAddress(pubKey2.Address()))
	val3 := hmTypes.NewValidator(3, 100, 0, 1, 100, hmTypes.NewPubKey(pubKey3[:]), hmTypes.IrisAddress(pubKey3.Address()))

	vals := []*hmTypes.Validator{val1, val2, val3}
	valSet := hmTypes.NewValidatorSet(vals)

	return valSet
}
