package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/cobra"

	"github.com/zenanetwork/iris/chainmanager/types"
	"github.com/zenanetwork/iris/version"
)

// GetQueryCmd returns the transaction commands for this module
func GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the chainmanager module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		client.GetCommands(
			GetQueryParams(cdc),
		)...,
	)

	return txCmd
}

// GetQueryParams implements the params query command.
func GetQueryParams(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "show the current chainmanager parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as chain manager parameters.

Example:
$ %s query chainmanager params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			if err = jsoniter.ConfigFastest.Unmarshal(bz, &params); err != nil {
				return err
			}
			return cliCtx.PrintOutput(params)
		},
	}
}
