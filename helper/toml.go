package helper

import (
	"bytes"
	"text/template"

	cmn "github.com/tendermint/tendermint/libs/common"
)

// Note: any changes to the comments/variables/mapstructure
// must be reflected in the appropriate struct in helper/config.go
const defaultConfigTemplate = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

##### RPC and REST configs #####

# RPC endpoint for ethereum chain
eth_rpc_url = "{{ .EthRPCUrl }}"

# RPC endpoint for zena chain
zena_rpc_url = "{{ .ZenaRPCUrl }}"

# GRPC flag for zena chain
zena_grpc_flag = "{{ .ZenaGRPCFlag }}"

# GRPC endpoint for zena chain
zena_grpc_url = "{{ .ZenaGRPCUrl }}"

# RPC endpoint for tendermint
tendermint_rpc_url = "{{ .TendermintRPCUrl }}"

# Polygon Sub Graph URL for self-heal mechanism (optional)
sub_graph_url = "{{ .SubGraphUrl }}"

#### Bridge configs ####

# Iris REST server endpoint, which is used by bridge
iris_rest_server = "{{ .IrisServerURL }}"

# AMQP endpoint
amqp_url = "{{ .AmqpURL }}"

## Poll intervals
checkpoint_poll_interval = "{{ .CheckpointerPollInterval }}"
syncer_poll_interval = "{{ .SyncerPollInterval }}"
noack_poll_interval = "{{ .NoACKPollInterval }}"
clerk_poll_interval = "{{ .ClerkPollInterval }}"
span_poll_interval = "{{ .SpanPollInterval }}"
milestone_poll_interval = "{{ .MilestonePollInterval }}"
enable_self_heal = "{{ .EnableSH }}"
sh_state_synced_interval = "{{ .SHStateSyncedInterval }}"
sh_stake_update_interval = "{{ .SHStakeUpdateInterval }}"
sh_max_depth_duration = "{{ .SHMaxDepthDuration }}"


#### gas limits ####
main_chain_gas_limit = "{{ .MainchainGasLimit }}"

#### gas price ####
main_chain_max_gas_price = "{{ .MainchainMaxGasPrice }}"

##### Timeout Config #####
no_ack_wait_time = "{{ .NoACKWaitTime }}"

##### chain - newSelectionAlgoHeight depends on this #####
chain = "{{ .Chain }}"
`

var configTemplate *template.Template

func init() {
	var err error

	tmpl := template.New("appConfigFileTemplate")
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// WriteConfigFile renders config using the template and writes it to
// configFilePath.
func WriteConfigFile(configFilePath string, config *Configuration) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, config); err != nil {
		panic(err)
	}

	cmn.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}
