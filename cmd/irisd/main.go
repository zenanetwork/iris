package main

import (
	"context"

	"github.com/maticnetwork/heimdall/cmd/heimdalld/service"
	"github.com/maticnetwork/heimdall/version"
)

func main() {
	version.UpdateIrisdInfo()
	service.NewIrisService(context.Background(), nil)
}
