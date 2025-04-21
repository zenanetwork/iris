package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	supplyTypes "github.com/zenanetwork/iris/supply/types"
	hmTypes "github.com/zenanetwork/iris/types"
)

// SupplyKeeper defines the supply Keeper for module accounts
type SupplyKeeper interface {
	GetModuleAddress(name string) hmTypes.IrisAddress
	GetModuleAccount(ctx sdk.Context, name string) supplyTypes.ModuleAccountInterface

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, supplyTypes.ModuleAccountInterface)

	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr hmTypes.IrisAddress, amt sdk.Coins) sdk.Error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr hmTypes.IrisAddress, recipientModule string, amt sdk.Coins) sdk.Error
}
