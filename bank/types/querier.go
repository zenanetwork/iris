package types

import (
	hmTyps "github.com/zenanetwork/iris/types"
)

const (
	QueryBalance = "balances"
)

// QueryBalanceParams defines the params for querying an account balance.
type QueryBalanceParams struct {
	Address hmTyps.IrisAddress
}

// NewQueryBalanceParams creates a new instance of QueryBalanceParams.
func NewQueryBalanceParams(addr hmTyps.IrisAddress) QueryBalanceParams {
	return QueryBalanceParams{Address: addr}
}
