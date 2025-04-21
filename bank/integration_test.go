package bank_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/zenanetwork/iris/app"
	authTypes "github.com/zenanetwork/iris/auth/types"
	bankTypes "github.com/zenanetwork/iris/bank/types"
)

//
// Create test app
//

// returns context and app with params set on account keeper
// nolint: unparam
func createTestApp(isCheckTx bool) (*app.IrisApp, sdk.Context) {
	app := app.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authTypes.DefaultParams())
	app.BankKeeper.SetSendEnabled(ctx, bankTypes.DefaultSendEnabled)

	return app, ctx
}
