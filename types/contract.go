package types

import (
	"github.com/zenanetwork/go-zenanet/accounts/abi"
	"github.com/zenanetwork/go-zenanet/accounts/abi/bind"
	"github.com/zenanetwork/go-zenanet/common"
)

// Contract is how we represent contracts at iris
type Contract struct {
	name    string
	address common.Address
	abi     abi.ABI
	// Location of contract
	// 0 - Ethereum Chain
	// 1 - Zena Chain
	location int
	instance bind.ContractBackend
}

// NewContract creates new contract instance
func NewContract(name string, address common.Address, abi abi.ABI, location int, instance bind.ContractBackend) Contract {
	return Contract{
		name:     name,
		address:  address,
		abi:      abi,
		location: location,
		instance: instance,
	}
}

// Location returns location of contract
func (c *Contract) Location() int {
	return c.location
}

// Name returns name of contract
func (c *Contract) Name() string {
	return c.name
}

// Address returns address of contract
func (c *Contract) Address() common.Address {
	return c.address
}

// ABI returns the abi of contract
func (c *Contract) ABI() abi.ABI {
	return c.abi
}

// Instance returns the instance of contract
func (c *Contract) Instance() bind.ContractBackend {
	return c.instance
}
