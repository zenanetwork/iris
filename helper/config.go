package helper

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	logger "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"
	tmTypes "github.com/tendermint/tendermint/types"
	ethCrypto "github.com/zenanetwork/go-zenanet/crypto"
	"github.com/zenanetwork/go-zenanet/ethclient"
	"github.com/zenanetwork/go-zenanet/rpc"

	"github.com/zenanetwork/iris/file"
	hmTypes "github.com/zenanetwork/iris/types"
	zenagrpc "github.com/zenanetwork/iris/zena/client/grpc"
)

const (
	TendermintNodeFlag   = "node"
	WithIrisConfigFlag   = "iris-config"
	HomeFlag             = "home"
	FlagClientHome       = "home-client"
	OverwriteGenesisFlag = "overwrite-genesis"
	RestServerFlag       = "rest-server"
	BridgeFlag           = "bridge"
	LogLevel             = "log_level"
	LogsWriterFileFlag   = "logs_writer_file"
	SeedsFlag            = "seeds"

	MainChain   = "mainnet"
	MumbaiChain = "mumbai"
	AmoyChain   = "amoy"
	LocalChain  = "local"

	// iris-config flags
	MainRPCUrlFlag               = "eth_rpc_url"
	ZenaRPCUrlFlag               = "zena_rpc_url"
	ZenaGRPCUrlFlag              = "zena_grpc_url"
	ZenaGRPCFlag                 = "zena_grpc_flag"
	TendermintNodeURLFlag        = "tendermint_rpc_url"
	IrisServerURLFlag            = "iris_rest_server"
	AmqpURLFlag                  = "amqp_url"
	CheckpointerPollIntervalFlag = "checkpoint_poll_interval"
	SyncerPollIntervalFlag       = "syncer_poll_interval"
	NoACKPollIntervalFlag        = "noack_poll_interval"
	ClerkPollIntervalFlag        = "clerk_poll_interval"
	SpanPollIntervalFlag         = "span_poll_interval"
	MilestonePollIntervalFlag    = "milestone_poll_interval"
	MainchainGasLimitFlag        = "main_chain_gas_limit"
	MainchainMaxGasPriceFlag     = "main_chain_max_gas_price"

	NoACKWaitTimeFlag = "no_ack_wait_time"
	ChainFlag         = "chain"

	// ---
	// TODO Move these to common client flags
	// BroadcastBlock defines a tx broadcasting mode where the client waits for
	// the tx to be committed in a block.
	BroadcastBlock = "block"

	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"

	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"
	// --

	// RPC Endpoints
	DefaultMainRPCUrl  = "http://localhost:9545"
	DefaultZenaRPCUrl  = "http://localhost:8545"
	DefaultZenaGRPCUrl = "localhost:3131"

	// RPC Timeouts
	DefaultEthRPCTimeout  = 5 * time.Second
	DefaultZenaRPCTimeout = 5 * time.Second

	// Services

	// DefaultAmqpURL represents default AMQP url
	DefaultAmqpURL           = "amqp://guest:guest@localhost:5672/"
	DefaultIrisServerURL     = "http://0.0.0.0:1317"
	DefaultTendermintNodeURL = "http://0.0.0.0:26657"

	NoACKWaitTime = 1800 * time.Second // Time ack service waits to clear buffer and elect new proposer (1800 seconds ~ 30 mins)

	DefaultCheckpointerPollInterval = 5 * time.Minute
	DefaultSyncerPollInterval       = 1 * time.Minute
	DefaultNoACKPollInterval        = 1010 * time.Second
	DefaultClerkPollInterval        = 10 * time.Second
	DefaultSpanPollInterval         = 1 * time.Minute

	DefaultMilestonePollInterval = 30 * time.Second

	DefaultEnableSH              = false
	DefaultSHStateSyncedInterval = 15 * time.Minute
	DefaultSHStakeUpdateInterval = 3 * time.Hour

	DefaultSHMaxDepthDuration = time.Hour

	DefaultMainchainGasLimit = uint64(5000000)

	DefaultMainchainMaxGasPrice = 400000000000 // 400 Gwei

	DefaultZenaChainID = "15001"

	DefaultLogsType = "json"
	DefaultChain    = MainChain

	DefaultTendermintNode = "tcp://localhost:26657"

	DefaultMainnetSeeds     = "e019e16d4e376723f3adc58eb1761809fea9bee0@35.234.150.253:26656,7f3049e88ac7f820fd86d9120506aaec0dc54b27@34.89.75.187:26656,1f5aff3b4f3193404423c3dd1797ce60cd9fea43@34.142.43.249:26656,2d5484feef4257e56ece025633a6ea132d8cadca@35.246.99.203:26656,17e9efcbd173e81a31579310c502e8cdd8b8ff2e@35.197.233.240:26656,72a83490309f9f63fdca3a0bef16c290e5cbb09c@35.246.95.65:26656,00677b1b2c6282fb060b7bb6e9cc7d2d05cdd599@34.105.180.11:26656,721dd4cebfc4b78760c7ee5d7b1b44d29a0aa854@34.147.169.102:26656,4760b3fc04648522a0bcb2d96a10aadee141ee89@34.89.55.74:26656"
	DefaultAmoyTestnetSeeds = "e4eabef3111155890156221f018b0ea3b8b64820@35.197.249.21:26656,811c3127677a4a34df907b021aad0c9d22f84bf4@34.89.39.114:26656,2ec15d1d33261e8cf42f57236fa93cfdc21c1cfb@35.242.167.175:26656,38120f9d2c003071a7230788da1e3129b6fb9d3f@34.89.15.223:26656,2f16f3857c6c99cc11e493c2082b744b8f36b127@34.105.128.110:26656,2833f06a5e33da2e80541fb1bfde2a7229877fcb@34.89.21.99:26656,2e6f1342416c5d758f5ae32f388bb76f7712a317@34.89.101.16:26656,a596f98b41851993c24de00a28b767c7c5ff8b42@34.89.11.233:26656"
	// Deprecated: Mumbai Testnet is deprecated
	DefaultMumbaiTestnetSeeds = "9df7ae4bf9b996c0e3436ed4cd3050dbc5742a28@43.200.206.40:26656,d9275750bc877b0276c374307f0fd7eae1d71e35@54.216.248.9:26656,1a3258eb2b69b235d4749cf9266a94567d6c0199@52.214.83.78:26656"

	secretFilePerm = 0600

	// Legacy value - DO NOT CHANGE
	// Maximum allowed event record data size
	LegacyMaxStateSyncSize = 100000

	// New max state sync size after hardfork
	MaxStateSyncSize = 30000

	//Milestone Length
	MilestoneLength = uint64(12)

	MilestonePruneNumber = uint64(100)

	MaticChainMilestoneConfirmation = uint64(16)

	//Milestone buffer Length
	MilestoneBufferLength = MilestoneLength * 5
	MilestoneBufferTime   = 256 * time.Second
	// Default Open Collector Endpoint
	DefaultOpenCollectorEndpoint = "localhost:4317"
)

var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.iriscli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.irisd")
	MinBalance      = big.NewInt(100000000000000000) // aka 0.1 Ether
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{}, secp256k1.PubKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PrivKeySecp256k1{}, secp256k1.PrivKeyAminoName, nil)

	Logger = logger.NewTMLogger(logger.NewSyncWriter(os.Stdout))
}

// Configuration represents iris config
type Configuration struct {
	EthRPCUrl        string `mapstructure:"eth_rpc_url"`        // RPC endpoint for main chain
	ZenaRPCUrl       string `mapstructure:"zena_rpc_url"`       // RPC endpoint for zena chain
	ZenaGRPCUrl      string `mapstructure:"zena_grpc_url"`      // gRPC endpoint for zena chain
	ZenaGRPCFlag     bool   `mapstructure:"zena_grpc_flag"`     // gRPC flag for zena chain
	TendermintRPCUrl string `mapstructure:"tendermint_rpc_url"` // tendemint node url
	SubGraphUrl      string `mapstructure:"sub_graph_url"`      // sub graph url

	EthRPCTimeout  time.Duration `mapstructure:"eth_rpc_timeout"`  // timeout for eth rpc
	ZenaRPCTimeout time.Duration `mapstructure:"zena_rpc_timeout"` // timeout for zena rpc

	AmqpURL       string `mapstructure:"amqp_url"`         // amqp url
	IrisServerURL string `mapstructure:"iris_rest_server"` // iris server url

	MainchainGasLimit uint64 `mapstructure:"main_chain_gas_limit"` // gas limit to mainchain transaction. eg....submit checkpoint.

	MainchainMaxGasPrice int64 `mapstructure:"main_chain_max_gas_price"` // max gas price to mainchain transaction. eg....submit checkpoint.

	// config related to bridge
	CheckpointerPollInterval time.Duration `mapstructure:"checkpoint_poll_interval"` // Poll interval for checkpointer service to send new checkpoints or missing ACK
	SyncerPollInterval       time.Duration `mapstructure:"syncer_poll_interval"`     // Poll interval for syncher service to sync for changes on main chain
	NoACKPollInterval        time.Duration `mapstructure:"noack_poll_interval"`      // Poll interval for ack service to send no-ack in case of no checkpoints
	ClerkPollInterval        time.Duration `mapstructure:"clerk_poll_interval"`
	SpanPollInterval         time.Duration `mapstructure:"span_poll_interval"`
	MilestonePollInterval    time.Duration `mapstructure:"milestone_poll_interval"`
	EnableSH                 bool          `mapstructure:"enable_self_heal"`         // Enable self healing
	SHStateSyncedInterval    time.Duration `mapstructure:"sh_state_synced_interval"` // Interval to self-heal StateSynced events if missing
	SHStakeUpdateInterval    time.Duration `mapstructure:"sh_stake_update_interval"` // Interval to self-heal StakeUpdate events if missing
	SHMaxDepthDuration       time.Duration `mapstructure:"sh_max_depth_duration"`    // Max duration that allows to suggest self-healing is not needed

	// wait time related options
	NoACKWaitTime time.Duration `mapstructure:"no_ack_wait_time"` // Time ack service waits to clear buffer and elect new proposer

	// Log related options
	LogsType       string `mapstructure:"logs_type"`        // if true, enable logging in json format
	LogsWriterFile string `mapstructure:"logs_writer_file"` // if given, Logs will be written to this file else os.Stdout

	// current chain - newSelectionAlgoHeight depends on this
	Chain string `mapstructure:"chain"`
}

var conf Configuration

// MainChainClient stores eth clie nt for Main chain Network
var mainChainClient *ethclient.Client
var mainRPCClient *rpc.Client

// MaticClient stores eth/rpc client for Matic Network
var maticClient *ethclient.Client
var maticRPCClient *rpc.Client
var maticGRPCClient *zenagrpc.ZenaGRPCClient

// private key object
var privObject secp256k1.PrivKeySecp256k1

var pubObject secp256k1.PubKeySecp256k1

// Logger stores global logger object
var Logger logger.Logger

// GenesisDoc contains the genesis file
var GenesisDoc tmTypes.GenesisDoc

var newSelectionAlgoHeight int64 = 0

var spanOverrideHeight int64 = 0

var milestoneZenaBlockHeight uint64 = 0

var aalzenagHeight int64 = 0

var newHexToStringAlgoHeight int64 = 0

var jorvikHeight int64 = 0

var danelawHeight int64 = 0

type ChainManagerAddressMigration struct {
	MaticTokenAddress     hmTypes.IrisAddress
	RootChainAddress      hmTypes.IrisAddress
	StakingManagerAddress hmTypes.IrisAddress
	SlashManagerAddress   hmTypes.IrisAddress
	StakingInfoAddress    hmTypes.IrisAddress
	StateSenderAddress    hmTypes.IrisAddress
}

var chainManagerAddressMigrations = map[string]map[int64]ChainManagerAddressMigration{
	MainChain:   {},
	MumbaiChain: {},
	AmoyChain:   {},
	"default":   {},
}

// Contracts
// var RootChain types.Contract
// var DepositManager types.Contract

// InitIrisConfig initializes with viper config (from iris configuration)
func InitIrisConfig(homeDir string) {
	if strings.Compare(homeDir, "") == 0 {
		// get home dir from viper
		homeDir = viper.GetString(HomeFlag)
	}

	// get iris config filepath from viper/cobra flag
	irisConfigFileFromFlag := viper.GetString(WithIrisConfigFlag)

	// init iris with changed config files
	InitIrisConfigWith(homeDir, irisConfigFileFromFlag)
}

// InitIrisConfigWith initializes passed iris/tendermint config files
func InitIrisConfigWith(homeDir string, irisConfigFileFromFLag string) {
	if strings.Compare(homeDir, "") == 0 {
		return
	}

	if strings.Compare(conf.ZenaRPCUrl, "") != 0 || strings.Compare(conf.ZenaGRPCUrl, "") != 0 {
		return
	}

	// read configuration from the standard configuration file
	configDir := filepath.Join(homeDir, "config")
	irisViper := viper.New()
	irisViper.SetEnvPrefix("IRIS")
	irisViper.AutomaticEnv()

	if irisConfigFileFromFLag == "" {
		irisViper.SetConfigName("iris-config") // name of config file (without extension)
		irisViper.AddConfigPath(configDir)     // call multiple times to add many search paths
	} else {
		irisViper.SetConfigFile(irisConfigFileFromFLag) // set config file explicitly
	}

	// Handle errors reading the config file
	if err := irisViper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// unmarshal configuration from the standard configuration file
	if err := irisViper.UnmarshalExact(&conf); err != nil {
		log.Fatalln("Unable to unmarshall config", "Error", err)
	}

	//  if there is a file with overrides submitted via flags => read it an merge it with the alreadey read standard configuration
	if irisConfigFileFromFLag != "" {
		irisViperFromFlag := viper.New()
		irisViperFromFlag.SetConfigFile(irisConfigFileFromFLag) // set flag config file explicitly

		err := irisViperFromFlag.ReadInConfig()
		if err != nil { // Handle errors reading the config file sybmitted as a flag
			log.Fatalln("Unable to read config file submitted via flag", "Error", err)
		}

		var confFromFlag Configuration
		// unmarshal configuration from the configuration file submitted as a flag
		if err = irisViperFromFlag.UnmarshalExact(&confFromFlag); err != nil {
			log.Fatalln("Unable to unmarshall config file submitted via flag", "Error", err)
		}

		conf.Merge(&confFromFlag)
	}

	// update configuration data with submitted flags
	if err := conf.UpdateWithFlags(viper.GetViper(), Logger); err != nil {
		log.Fatalln("Unable to read flag values. Check log for details.", "Error", err)
	}

	// perform check for json logging
	if conf.LogsType == "json" {
		Logger = logger.NewTMJSONLogger(logger.NewSyncWriter(GetLogsWriter(conf.LogsWriterFile)))
	} else {
		// default fallback
		Logger = logger.NewTMLogger(logger.NewSyncWriter(GetLogsWriter(conf.LogsWriterFile)))
	}

	// perform checks for timeout
	if conf.EthRPCTimeout == 0 {
		// fallback to default
		Logger.Debug("Missing ETH RPC timeout or invalid value provided, falling back to default", "timeout", DefaultEthRPCTimeout)
		conf.EthRPCTimeout = DefaultEthRPCTimeout
	}

	if conf.ZenaRPCTimeout == 0 {
		// fallback to default
		Logger.Debug("Missing ZENA RPC timeout or invalid value provided, falling back to default", "timeout", DefaultZenaRPCTimeout)
		conf.ZenaRPCTimeout = DefaultZenaRPCTimeout
	}

	if conf.SHStateSyncedInterval == 0 {
		// fallback to default
		Logger.Debug("Missing self-healing StateSynced interval or invalid value provided, falling back to default", "interval", DefaultSHStateSyncedInterval)
		conf.SHStateSyncedInterval = DefaultSHStateSyncedInterval
	}

	if conf.SHStakeUpdateInterval == 0 {
		// fallback to default
		Logger.Debug("Missing self-healing StakeUpdate interval or invalid value provided, falling back to default", "interval", DefaultSHStakeUpdateInterval)
		conf.SHStakeUpdateInterval = DefaultSHStakeUpdateInterval
	}

	if conf.SHMaxDepthDuration == 0 {
		// fallback to default
		Logger.Debug("Missing self-healing max depth duration or invalid value provided, falling back to default", "duration", DefaultSHMaxDepthDuration)
		conf.SHMaxDepthDuration = DefaultSHMaxDepthDuration
	}

	var err error
	if mainRPCClient, err = rpc.Dial(conf.EthRPCUrl); err != nil {
		log.Fatalln("Unable to dial via ethClient", "URL=", conf.EthRPCUrl, "chain=eth", "Error", err)
	}

	mainChainClient = ethclient.NewClient(mainRPCClient)

	if maticRPCClient, err = rpc.Dial(conf.ZenaRPCUrl); err != nil {
		log.Fatal(err)
	}

	maticClient = ethclient.NewClient(maticRPCClient)

	maticGRPCClient = zenagrpc.NewZenaGRPCClient(conf.ZenaGRPCUrl)

	// Loading genesis doc
	genDoc, err := tmTypes.GenesisDocFromFile(filepath.Join(configDir, "genesis.json"))
	if err != nil {
		log.Fatal(err)
	}

	GenesisDoc = *genDoc

	// load pv file, unmarshall and set to privObject
	err = file.PermCheck(file.Rootify("priv_validator_key.json", configDir), secretFilePerm)
	if err != nil {
		Logger.Error(err.Error())
	}

	privVal := privval.LoadFilePV(filepath.Join(configDir, "priv_validator_key.json"), filepath.Join(configDir, "priv_validator_key.json"))
	cdc.MustUnmarshalBinaryBare(privVal.Key.PrivKey.Bytes(), &privObject)
	cdc.MustUnmarshalBinaryBare(privObject.PubKey().Bytes(), &pubObject)

	switch conf.Chain {
	case MainChain:
		newSelectionAlgoHeight = 375300
		spanOverrideHeight = 8664000
		newHexToStringAlgoHeight = 9266260
		aalzenagHeight = 15950759
		jorvikHeight = 22393043
		danelawHeight = 22393043
	case MumbaiChain:
		newSelectionAlgoHeight = 282500
		spanOverrideHeight = 10205000
		newHexToStringAlgoHeight = 12048023
		aalzenagHeight = 18035772
		jorvikHeight = -1
		danelawHeight = -1
	case AmoyChain:
		newSelectionAlgoHeight = 0
		spanOverrideHeight = 0
		newHexToStringAlgoHeight = 0
		aalzenagHeight = 0
		jorvikHeight = 5768528
		danelawHeight = 6490424
	default:
		newSelectionAlgoHeight = 0
		spanOverrideHeight = 0
		newHexToStringAlgoHeight = 0
		aalzenagHeight = 0
		jorvikHeight = 0
		danelawHeight = 0
	}
}

// GetDefaultIrisConfig returns configuration with default params
func GetDefaultIrisConfig() Configuration {
	return Configuration{
		EthRPCUrl:        DefaultMainRPCUrl,
		ZenaRPCUrl:       DefaultZenaRPCUrl,
		ZenaGRPCUrl:      DefaultZenaGRPCUrl,
		TendermintRPCUrl: DefaultTendermintNodeURL,

		EthRPCTimeout:  DefaultEthRPCTimeout,
		ZenaRPCTimeout: DefaultZenaRPCTimeout,

		AmqpURL:       DefaultAmqpURL,
		IrisServerURL: DefaultIrisServerURL,

		MainchainGasLimit: DefaultMainchainGasLimit,

		MainchainMaxGasPrice: DefaultMainchainMaxGasPrice,

		CheckpointerPollInterval: DefaultCheckpointerPollInterval,
		SyncerPollInterval:       DefaultSyncerPollInterval,
		NoACKPollInterval:        DefaultNoACKPollInterval,
		ClerkPollInterval:        DefaultClerkPollInterval,
		SpanPollInterval:         DefaultSpanPollInterval,
		MilestonePollInterval:    DefaultMilestonePollInterval,
		EnableSH:                 DefaultEnableSH,
		SHStateSyncedInterval:    DefaultSHStateSyncedInterval,
		SHStakeUpdateInterval:    DefaultSHStakeUpdateInterval,
		SHMaxDepthDuration:       DefaultSHMaxDepthDuration,

		NoACKWaitTime: NoACKWaitTime,

		LogsType:       DefaultLogsType,
		Chain:          DefaultChain,
		LogsWriterFile: "", // default to stdout
	}
}

// GetConfig returns cached configuration object
func GetConfig() Configuration {
	return conf
}

func GetGenesisDoc() tmTypes.GenesisDoc {
	return GenesisDoc
}

// TEST PURPOSE ONLY
// SetTestConfig sets test configuration
func SetTestConfig(_conf Configuration) {
	conf = _conf
}

// TEST PURPOSE ONLY
// SetTestPrivPubKey sets test priv and pub key for testing
func SetTestPrivPubKey(privKey secp256k1.PrivKeySecp256k1) {
	privObject = privKey
	privObject.PubKey()
	pubKey, ok := privObject.PubKey().(secp256k1.PubKeySecp256k1)
	if !ok {
		panic("pub key is not of type secp256k1.PubKeySecp256k1")
	}
	pubObject = pubKey
}

//
// Get main/matic clients
//

// GetMainChainRPCClient returns main chain RPC client
func GetMainChainRPCClient() *rpc.Client {
	return mainRPCClient
}

// GetMainClient returns main chain's eth client
func GetMainClient() *ethclient.Client {
	return mainChainClient
}

// GetMaticClient returns matic's eth client
func GetMaticClient() *ethclient.Client {
	return maticClient
}

// GetMaticRPCClient returns matic's RPC client
func GetMaticRPCClient() *rpc.Client {
	return maticRPCClient
}

// GetMaticGRPCClient returns matic's gRPC client
func GetMaticGRPCClient() *zenagrpc.ZenaGRPCClient {
	return maticGRPCClient
}

// GetPrivKey returns priv key object
func GetPrivKey() secp256k1.PrivKeySecp256k1 {
	return privObject
}

// GetECDSAPrivKey return ecdsa private key
func GetECDSAPrivKey() *ecdsa.PrivateKey {
	// get priv key
	pkObject := GetPrivKey()

	// create ecdsa private key
	ecdsaPrivateKey, _ := ethCrypto.ToECDSA(pkObject[:])

	return ecdsaPrivateKey
}

// GetPubKey returns pub key object
func GetPubKey() secp256k1.PubKeySecp256k1 {
	return pubObject
}

// GetAddress returns address object
func GetAddress() []byte {
	return GetPubKey().Address().Bytes()
}

// GetValidChains returns all the valid chains
func GetValidChains() []string {
	return []string{"mainnet", "mumbai", "amoy", "local"}
}

// GetNewSelectionAlgoHeight returns newSelectionAlgoHeight
func GetNewSelectionAlgoHeight() int64 {
	return newSelectionAlgoHeight
}

// GetSpanOverrideHeight returns spanOverrideHeight
func GetSpanOverrideHeight() int64 {
	return spanOverrideHeight
}

// GetAalborgHardForkHeight returns AalzenagHardForkHeight
func GetAalborgHardForkHeight() int64 {
	return aalzenagHeight
}

// GetMilestoneZenaBlockHeight returns milestoneZenaBlockHeight
func GetMilestoneZenaBlockHeight() uint64 {
	return milestoneZenaBlockHeight
}

// GetNewHexToStringAlgoHeight returns newHexToStringAlgoHeight
func GetNewHexToStringAlgoHeight() int64 {
	return newHexToStringAlgoHeight
}

// GetJorvikHeight returns jorvikHeight
func GetJorvikHeight() int64 {
	return jorvikHeight
}

// GetDanelawHeight returns danelawHeight
func GetDanelawHeight() int64 {
	return danelawHeight
}

func GetChainManagerAddressMigration(blockNum int64) (ChainManagerAddressMigration, bool) {
	chainMigration := chainManagerAddressMigrations[conf.Chain]
	if chainMigration == nil {
		chainMigration = chainManagerAddressMigrations["default"]
	}

	result, found := chainMigration[blockNum]

	return result, found
}

// DecorateWithIrisFlags adds persistent flags for iris-config and bind flags with command
func DecorateWithIrisFlags(cmd *cobra.Command, v *viper.Viper, loggerInstance logger.Logger, caller string) {
	// add with-iris-config flag
	cmd.PersistentFlags().String(
		WithIrisConfigFlag,
		"",
		"Override of Iris config file (default <home>/config/iris-config.json)",
	)

	if err := v.BindPFlag(WithIrisConfigFlag, cmd.PersistentFlags().Lookup(WithIrisConfigFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, WithIrisConfigFlag), "Error", err)
	}

	// add MainRPCUrlFlag flag
	cmd.PersistentFlags().String(
		MainRPCUrlFlag,
		"",
		"Set RPC endpoint for ethereum chain",
	)

	if err := v.BindPFlag(MainRPCUrlFlag, cmd.PersistentFlags().Lookup(MainRPCUrlFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, MainRPCUrlFlag), "Error", err)
	}

	// add ZenaRPCUrlFlag flag
	cmd.PersistentFlags().String(
		ZenaRPCUrlFlag,
		"",
		"Set RPC endpoint for zena chain",
	)

	if err := v.BindPFlag(ZenaRPCUrlFlag, cmd.PersistentFlags().Lookup(ZenaRPCUrlFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, ZenaRPCUrlFlag), "Error", err)
	}

	// add ZenaGRPCUrlFlag flag
	cmd.PersistentFlags().String(
		ZenaGRPCUrlFlag,
		"",
		"Set gRPC endpoint for zena chain",
	)

	if err := v.BindPFlag(ZenaGRPCUrlFlag, cmd.PersistentFlags().Lookup(ZenaGRPCUrlFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, ZenaGRPCUrlFlag), "Error", err)
	}

	cmd.PersistentFlags().Bool(
		ZenaGRPCFlag,
		false,
		"Set if iris will use gRPC or Rest to interact with zena chain",
	)

	if err := v.BindPFlag(ZenaGRPCFlag, cmd.PersistentFlags().Lookup(ZenaGRPCFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, ZenaGRPCFlag), "Error", err)
	}

	// add TendermintNodeURLFlag flag
	cmd.PersistentFlags().String(
		TendermintNodeURLFlag,
		"",
		"Set RPC endpoint for tendermint",
	)

	if err := v.BindPFlag(TendermintNodeURLFlag, cmd.PersistentFlags().Lookup(TendermintNodeURLFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, TendermintNodeURLFlag), "Error", err)
	}

	// add IrisServerURLFlag flag
	cmd.PersistentFlags().String(
		IrisServerURLFlag,
		"",
		"Set Iris REST server endpoint",
	)

	if err := v.BindPFlag(IrisServerURLFlag, cmd.PersistentFlags().Lookup(IrisServerURLFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, IrisServerURLFlag), "Error", err)
	}

	// add AmqpURLFlag flag
	cmd.PersistentFlags().String(
		AmqpURLFlag,
		"",
		"Set AMQP endpoint",
	)

	if err := v.BindPFlag(AmqpURLFlag, cmd.PersistentFlags().Lookup(AmqpURLFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, AmqpURLFlag), "Error", err)
	}

	// add CheckpointerPollIntervalFlag flag
	cmd.PersistentFlags().String(
		CheckpointerPollIntervalFlag,
		"",
		"Set check point pull interval",
	)

	if err := v.BindPFlag(CheckpointerPollIntervalFlag, cmd.PersistentFlags().Lookup(CheckpointerPollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, CheckpointerPollIntervalFlag), "Error", err)
	}

	// add SyncerPollIntervalFlag flag
	cmd.PersistentFlags().String(
		SyncerPollIntervalFlag,
		"",
		"Set syncer pull interval",
	)

	if err := v.BindPFlag(SyncerPollIntervalFlag, cmd.PersistentFlags().Lookup(SyncerPollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, SyncerPollIntervalFlag), "Error", err)
	}

	// add NoACKPollIntervalFlag flag
	cmd.PersistentFlags().String(
		NoACKPollIntervalFlag,
		"",
		"Set no acknowledge pull interval",
	)

	if err := v.BindPFlag(NoACKPollIntervalFlag, cmd.PersistentFlags().Lookup(NoACKPollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, NoACKPollIntervalFlag), "Error", err)
	}

	// add ClerkPollIntervalFlag flag
	cmd.PersistentFlags().String(
		ClerkPollIntervalFlag,
		"",
		"Set clerk pull interval",
	)

	if err := v.BindPFlag(ClerkPollIntervalFlag, cmd.PersistentFlags().Lookup(ClerkPollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, ClerkPollIntervalFlag), "Error", err)
	}

	// add SpanPollIntervalFlag flag
	cmd.PersistentFlags().String(
		SpanPollIntervalFlag,
		"",
		"Set span pull interval",
	)

	if err := v.BindPFlag(SpanPollIntervalFlag, cmd.PersistentFlags().Lookup(SpanPollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, SpanPollIntervalFlag), "Error", err)
	}

	// add MilestonePollIntervalFlag flag
	cmd.PersistentFlags().String(
		MilestonePollIntervalFlag,
		DefaultMilestonePollInterval.String(),
		"Set milestone interval",
	)

	if err := v.BindPFlag(MilestonePollIntervalFlag, cmd.PersistentFlags().Lookup(MilestonePollIntervalFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, MilestonePollIntervalFlag), "Error", err)
	}

	// add MainchainGasLimitFlag flag
	cmd.PersistentFlags().Uint64(
		MainchainGasLimitFlag,
		0,
		"Set main chain gas limit",
	)

	if err := v.BindPFlag(MainchainGasLimitFlag, cmd.PersistentFlags().Lookup(MainchainGasLimitFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, MainchainGasLimitFlag), "Error", err)
	}

	// add MainchainMaxGasPriceFlag flag
	cmd.PersistentFlags().Int64(
		MainchainMaxGasPriceFlag,
		0,
		"Set main chain max gas limit",
	)

	if err := v.BindPFlag(MainchainMaxGasPriceFlag, cmd.PersistentFlags().Lookup(MainchainMaxGasPriceFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, MainchainMaxGasPriceFlag), "Error", err)
	}

	// add NoACKWaitTimeFlag flag
	cmd.PersistentFlags().String(
		NoACKWaitTimeFlag,
		"",
		"Set time ack service waits to clear buffer and elect new proposer",
	)

	if err := v.BindPFlag(NoACKWaitTimeFlag, cmd.PersistentFlags().Lookup(NoACKWaitTimeFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, NoACKWaitTimeFlag), "Error", err)
	}

	// add chain flag
	cmd.PersistentFlags().String(
		ChainFlag,
		"",
		fmt.Sprintf("Set one of the chains: [%s]", strings.Join(GetValidChains(), ",")),
	)

	if err := v.BindPFlag(ChainFlag, cmd.PersistentFlags().Lookup(ChainFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, ChainFlag), "Error", err)
	}

	// add logsWriterFile flag
	cmd.PersistentFlags().String(
		LogsWriterFileFlag,
		"",
		"Set logs writer file, Default is os.Stdout",
	)

	if err := v.BindPFlag(LogsWriterFileFlag, cmd.PersistentFlags().Lookup(LogsWriterFileFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", caller, LogsWriterFileFlag), "Error", err)
	}
}

func (c *Configuration) UpdateWithFlags(v *viper.Viper, loggerInstance logger.Logger) error {
	const logErrMsg = "Unable to read flag."

	// get endpoint for ethereum chain from viper/cobra
	stringConfgValue := v.GetString(MainRPCUrlFlag)
	if stringConfgValue != "" {
		c.EthRPCUrl = stringConfgValue
	}

	// get endpoint for zena chain from viper/cobra
	stringConfgValue = v.GetString(ZenaRPCUrlFlag)
	if stringConfgValue != "" {
		c.ZenaRPCUrl = stringConfgValue
	}

	// get endpoint for zena chain from viper/cobra
	stringConfgValue = v.GetString(ZenaGRPCUrlFlag)
	if stringConfgValue != "" {
		c.ZenaGRPCUrl = stringConfgValue
	}

	// get gRPC flag for zena chain from viper/cobra
	boolConfgValue := v.GetBool(ZenaGRPCFlag)
	if boolConfgValue {
		c.ZenaGRPCFlag = boolConfgValue
	}

	// get endpoint for tendermint from viper/cobra
	stringConfgValue = v.GetString(TendermintNodeURLFlag)
	if stringConfgValue != "" {
		c.TendermintRPCUrl = stringConfgValue
	}

	// get endpoint for tendermint from viper/cobra
	stringConfgValue = v.GetString(AmqpURLFlag)
	if stringConfgValue != "" {
		c.AmqpURL = stringConfgValue
	}

	// get Iris REST server endpoint from viper/cobra
	stringConfgValue = v.GetString(IrisServerURLFlag)
	if stringConfgValue != "" {
		c.IrisServerURL = stringConfgValue
	}

	// need this error for parsing Duration values
	var err error

	// get check point pull interval from viper/cobra
	stringConfgValue = v.GetString(CheckpointerPollIntervalFlag)
	if stringConfgValue != "" {
		if c.CheckpointerPollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", CheckpointerPollIntervalFlag, "Error", err)
			return err
		}
	}

	// get syncer pull interval from viper/cobra
	stringConfgValue = v.GetString(SyncerPollIntervalFlag)
	if stringConfgValue != "" {
		if c.SyncerPollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", SyncerPollIntervalFlag, "Error", err)
			return err
		}
	}

	// get poll interval for ack service to send no-ack in case of no checkpoints from viper/cobra
	stringConfgValue = v.GetString(NoACKPollIntervalFlag)
	if stringConfgValue != "" {
		if c.NoACKPollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", NoACKPollIntervalFlag, "Error", err)
			return err
		}
	}

	// get clerk poll interval from viper/cobra
	stringConfgValue = v.GetString(ClerkPollIntervalFlag)
	if stringConfgValue != "" {
		if c.ClerkPollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", ClerkPollIntervalFlag, "Error", err)
			return err
		}
	}

	// get span poll interval from viper/cobra
	stringConfgValue = v.GetString(SpanPollIntervalFlag)
	if stringConfgValue != "" {
		if c.SpanPollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", SpanPollIntervalFlag, "Error", err)
			return err
		}
	}

	// get milestone poll interval from viper/cobra
	stringConfgValue = v.GetString(MilestonePollIntervalFlag)
	if stringConfgValue != "" {
		if c.MilestonePollInterval, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", MilestonePollIntervalFlag, "Error", err)
			return err
		}
	}

	// get time that ack service waits to clear buffer and elect new proposer from viper/cobra
	stringConfgValue = v.GetString(NoACKWaitTimeFlag)
	if stringConfgValue != "" {
		if c.NoACKWaitTime, err = time.ParseDuration(stringConfgValue); err != nil {
			loggerInstance.Error(logErrMsg, "Flag", NoACKWaitTimeFlag, "Error", err)
			return err
		}
	}

	// get mainchain gas limit from viper/cobra
	uint64ConfgValue := v.GetUint64(MainchainGasLimitFlag)
	if uint64ConfgValue != 0 {
		c.MainchainGasLimit = uint64ConfgValue
	}

	// get mainchain max gas price from viper/cobra. if it is greater then  zero => set it as configuration parameter
	int64ConfgValue := v.GetInt64(MainchainMaxGasPriceFlag)
	if int64ConfgValue > 0 {
		c.MainchainMaxGasPrice = int64ConfgValue
	}

	// get chain from viper/cobra flag
	stringConfgValue = v.GetString(ChainFlag)
	if stringConfgValue != "" {
		c.Chain = stringConfgValue
	}

	stringConfgValue = v.GetString(LogsWriterFileFlag)
	if stringConfgValue != "" {
		c.LogsWriterFile = stringConfgValue
	}

	return nil
}

func (c *Configuration) Merge(cc *Configuration) {
	if cc.EthRPCUrl != "" {
		c.EthRPCUrl = cc.EthRPCUrl
	}

	if cc.ZenaRPCUrl != "" {
		c.ZenaRPCUrl = cc.ZenaRPCUrl
	}

	if cc.ZenaGRPCUrl != "" {
		c.ZenaGRPCUrl = cc.ZenaGRPCUrl
	}

	if cc.TendermintRPCUrl != "" {
		c.TendermintRPCUrl = cc.TendermintRPCUrl
	}

	if cc.AmqpURL != "" {
		c.AmqpURL = cc.AmqpURL
	}

	if cc.IrisServerURL != "" {
		c.IrisServerURL = cc.IrisServerURL
	}

	if cc.MainchainGasLimit != 0 {
		c.MainchainGasLimit = cc.MainchainGasLimit
	}

	if cc.MainchainMaxGasPrice != 0 {
		c.MainchainMaxGasPrice = cc.MainchainMaxGasPrice
	}

	if cc.CheckpointerPollInterval != 0 {
		c.CheckpointerPollInterval = cc.CheckpointerPollInterval
	}

	if cc.SyncerPollInterval != 0 {
		c.SyncerPollInterval = cc.SyncerPollInterval
	}

	if cc.NoACKPollInterval != 0 {
		c.NoACKPollInterval = cc.NoACKPollInterval
	}

	if cc.ClerkPollInterval != 0 {
		c.ClerkPollInterval = cc.ClerkPollInterval
	}

	if cc.SpanPollInterval != 0 {
		c.SpanPollInterval = cc.SpanPollInterval
	}

	if cc.MilestonePollInterval != 0 {
		c.MilestonePollInterval = cc.MilestonePollInterval
	}

	if cc.NoACKWaitTime != 0 {
		c.NoACKWaitTime = cc.NoACKWaitTime
	}

	if cc.Chain != "" {
		c.Chain = cc.Chain
	}

	if cc.LogsWriterFile != "" {
		c.LogsWriterFile = cc.LogsWriterFile
	}
}

// DecorateWithTendermintFlags creates tendermint flags for desired command and bind them to viper
func DecorateWithTendermintFlags(cmd *cobra.Command, v *viper.Viper, loggerInstance logger.Logger, message string) {
	// add seeds flag
	cmd.PersistentFlags().String(
		SeedsFlag,
		"",
		"Override seeds",
	)

	if err := v.BindPFlag(SeedsFlag, cmd.PersistentFlags().Lookup(SeedsFlag)); err != nil {
		loggerInstance.Error(fmt.Sprintf("%v | BindPFlag | %v", message, SeedsFlag), "Error", err)
	}
}

// UpdateTendermintConfig updates tenedermint config with flags and default values if needed
func UpdateTendermintConfig(tendermintConfig *cfg.Config, v *viper.Viper) {
	// update tendermintConfig.P2P.Seeds
	seedsFlagValue := v.GetString(SeedsFlag)
	if seedsFlagValue != "" {
		tendermintConfig.P2P.Seeds = seedsFlagValue
	}

	if tendermintConfig.P2P.Seeds == "" {
		switch conf.Chain {
		case MainChain:
			tendermintConfig.P2P.Seeds = DefaultMainnetSeeds
		case MumbaiChain:
			tendermintConfig.P2P.Seeds = DefaultMumbaiTestnetSeeds
		case AmoyChain:
			tendermintConfig.P2P.Seeds = DefaultAmoyTestnetSeeds
		}
	}
}

func GetLogsWriter(logsWriterFile string) io.Writer {
	if logsWriterFile != "" {
		logWriter, err := os.OpenFile(logsWriterFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening log writer file: %v", err)
		}

		return logWriter
	} else {
		return os.Stdout
	}
}
