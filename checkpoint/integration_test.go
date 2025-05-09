package checkpoint_test

import (
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/zenanetwork/iris/app"
	"github.com/zenanetwork/iris/checkpoint/types"
	"github.com/zenanetwork/iris/helper"
	hmTypes "github.com/zenanetwork/iris/types"
)

//
// Create test app
//

// createTestApp returns context and app
func createTestApp(isCheckTx bool) (*app.IrisApp, sdk.Context, context.CLIContext) {
	genesisState := app.NewDefaultGenesisState()

	app := app.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	cliCtx := context.NewCLIContext().WithCodec(app.Codec())

	helper.SetTestConfig(helper.GetDefaultIrisConfig())

	params := types.NewParams(5*time.Second, 256, 1024, 10000)

	Checkpoints := make([]hmTypes.Checkpoint, 0)

	for i := range Checkpoints {
		Checkpoints[i] = hmTypes.Checkpoint{}
	}

	checkpointGenesis := types.NewGenesisState(
		types.DefaultGenesisState().Params,
		types.DefaultGenesisState().BufferedCheckpoint,
		types.DefaultGenesisState().LastNoACK,
		types.DefaultGenesisState().AckCount,
		types.DefaultGenesisState().Checkpoints,
	)

	genesisState[types.ModuleName] = app.Codec().MustMarshalJSON(checkpointGenesis)

	stateBytes, err := codec.MarshalJSONIndent(app.Codec(), genesisState)
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abci.RequestInitChain{
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: app.LastBlockHeight() + 1}})
	app.CheckpointKeeper.SetParams(ctx, params)

	return app, ctx, cliCtx
}
