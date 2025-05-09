package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/zenanetwork/iris/bridge/setu/util"
	clerkTypes "github.com/zenanetwork/iris/clerk/types"
	hmClient "github.com/zenanetwork/iris/client"
	"github.com/zenanetwork/iris/helper"
	"github.com/zenanetwork/iris/types"
	hmTypes "github.com/zenanetwork/iris/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        clerkTypes.ModuleName,
		Short:                      "Checkpoint transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       hmClient.ValidateCmd,
	}

	txCmd.AddCommand(
		client.PostCommands(
			CreateNewStateRecord(cdc),
		)...,
	)

	return txCmd
}

// CreateNewStateRecord send checkpoint transaction
func CreateNewStateRecord(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "new state record",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// zena chain id
			zenaChainID := viper.GetString(FlagZenaChainId)
			if zenaChainID == "" {
				return fmt.Errorf("ZenaChainID cannot be empty")
			}

			// get proposer
			proposer := types.HexToIrisAddress(viper.GetString(FlagProposerAddress))
			if proposer.Empty() {
				proposer = helper.GetFromAddress(cliCtx)
			}

			// tx hash
			txHashStr := viper.GetString(FlagTxHash)
			if txHashStr == "" {
				return fmt.Errorf("tx hash cannot be empty")
			}

			// tx hash
			recordIDStr := viper.GetString(FlagRecordID)
			if recordIDStr == "" {
				return fmt.Errorf("record id cannot be empty")
			}

			recordID, err := strconv.ParseUint(recordIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("record id cannot be empty")
			}

			// get contract Addr
			contractAddr := types.HexToIrisAddress(viper.GetString(FlagContractAddress))
			if contractAddr.Empty() {
				return fmt.Errorf("contract Address cannot be empty")
			}

			// log index
			logIndexStr := viper.GetString(FlagLogIndex)
			if logIndexStr == "" {
				return fmt.Errorf("log index cannot be empty")
			}

			logIndex, err := strconv.ParseUint(logIndexStr, 10, 64)
			if err != nil {
				return fmt.Errorf("log index cannot be parsed")
			}

			// log index
			dataStr := viper.GetString(FlagData)
			if dataStr == "" {
				return fmt.Errorf("data cannot be empty")
			}

			data := types.HexToHexBytes(dataStr)
			if dataStr == "" {
				return fmt.Errorf("data should be hex string")
			}

			if util.GetBlockHeight(cliCtx) > helper.GetSpanOverrideHeight() && len(data) > helper.MaxStateSyncSize {
				logger.Info(`Data is too large to process, Resetting to ""`, "id", recordIDStr)
				data = hmTypes.HexToHexBytes("")
			} else if len(data) > helper.LegacyMaxStateSyncSize {
				logger.Info(`Data is too large to process, Resetting to ""`, "id", recordIDStr)
				data = hmTypes.HexToHexBytes("")
			}

			// create new state record
			msg := clerkTypes.NewMsgEventRecord(
				proposer,
				types.HexToIrisHash(txHashStr),
				logIndex,
				viper.GetUint64(FlagBlockNumber),
				recordID,
				contractAddr,
				data,
				zenaChainID,
			)

			return helper.BroadcastMsgsWithCLI(cliCtx, []sdk.Msg{msg})
		},
	}
	cmd.Flags().StringP(FlagProposerAddress, "p", "", "--proposer=<proposer-address>")
	cmd.Flags().String(FlagTxHash, "", "--tx-hash=<tx-hash>")
	cmd.Flags().String(FlagLogIndex, "", "--log-index=<log-index>")
	cmd.Flags().String(FlagRecordID, "", "--id=<record-id>")
	cmd.Flags().String(FlagZenaChainId, "", "--zena-chain-id=<zena-chain-id>")
	cmd.Flags().Uint64(FlagBlockNumber, 0, "--block-number=<block-number>")
	cmd.Flags().String(FlagContractAddress, "", "--contract-addr=<contract-addr>")
	cmd.Flags().String(FlagData, "", "--data=<data>")

	if err := cmd.MarkFlagRequired(FlagRecordID); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagRecordID", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagTxHash); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagTxHash", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagLogIndex); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagLogIndex", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagZenaChainId); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagZenaChainId", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagBlockNumber); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagBlockNumber", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagContractAddress); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagContractAddress", "Error", err)
	}

	if err := cmd.MarkFlagRequired(FlagData); err != nil {
		logger.Error("CreateNewStateRecord | MarkFlagRequired | FlagData", "Error", err)
	}

	return cmd
}
