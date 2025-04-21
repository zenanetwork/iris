package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "clerk"

	// StoreKey is the store key string for zena
	StoreKey = ModuleName

	// RouterKey is the message route for zena
	RouterKey = ModuleName

	// QuerierRoute is the querier route for zena
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName

	// DefaultCodespace default code space
	DefaultCodespace sdk.CodespaceType = ModuleName
)
