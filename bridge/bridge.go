package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/zenanetwork/iris/bridge/cmd"
	"github.com/zenanetwork/iris/helper"
)

func main() {
	var logger = helper.Logger.With("module", "bridge/cmd/")
	rootCmd := cmd.BridgeCommands(viper.GetViper(), logger, "bridge-main")

	// add iris flags
	helper.DecorateWithIrisFlags(rootCmd, viper.GetViper(), logger, "bridge-main")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
