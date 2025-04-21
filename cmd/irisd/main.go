package main

import (
	"context"

	"github.com/zenanetwork/iris/cmd/irisd/service"
	"github.com/zenanetwork/iris/version"
)

func main() {
	version.UpdateIrisdInfo()
	service.NewIrisService(context.Background(), nil)
}
