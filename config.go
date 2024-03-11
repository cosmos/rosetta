package rosetta

import (
	"fmt"
	"time"

	crgerrs "github.com/cosmos/rosetta/lib/errors"

	"github.com/coinbase/rosetta-sdk-go/types"
	url "github.com/goware/urlx"
	"github.com/spf13/pflag"

	crg "github.com/cosmos/rosetta/lib/server"

	clientflags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	HTTP  = "http"
	HTTPS = "https"
	TCP   = "tcp"

	HTTPSPORT = "443"
)

// configuration defaults constants
const (
	// DefaultBlockchain defines the default blockchain identifier name
	DefaultBlockchain = "app"
	// DefaultAddr defines the default rosetta binding address
	DefaultAddr = ":8080"
	// DefaultRetries is the default number of retries
	DefaultRetries = 5
	// DefaultCometEndpoint is the default value for the CometBFT endpoint
	DefaultCometEndpoint = "localhost:26657"
	// DefaultGRPCEndpoint is the default value for the gRPC endpoint
	DefaultGRPCEndpoint = "localhost:9090"
	// DefaultGRPCEndpoint is the default value for the gRPC endpoint
	DefaultGRPCTypesServerEndpoint = ""
	// DefaultNetwork defines the default network name
	DefaultNetwork = "network"
	// DefaultOffline defines the default offline value
	DefaultOffline = false
	// DefaultEnableFeeSuggestion indicates to use fee suggestion if `construction/metadata` is called without gas limit and price
	DefaultEnableFeeSuggestion = false
	// DenomToSuggest defines the default denom for fee suggestion
	DenomToSuggest = "uatom"
	// DefaultPrices defines the default list of prices to suggest
	DefaultPrices = "1uatom,1stake"
	// DefaultPlugin define plugin location for interface and type registry
	DefaultPlugin = "cosmos-hub"
)

// configuration flags
const (
	FlagBlockchain              = "blockchain"
	FlagNetwork                 = "network"
	FlagTendermintEndpoint      = "tendermint"
	FlagGRPCEndpoint            = "grpc"
	FlagGRPCTypesServerEndpoint = "grpc-types-server"
	FlagAddr                    = "addr"
	FlagRetries                 = "retries"
	FlagOffline                 = "offline"
	FlagEnableFeeSuggestion     = "enable-fee-suggestion"
	FlagGasToSuggest            = "gas-to-suggest"
	FlagDenomToSuggest          = "denom-to-suggest"
	FlagPricesToSuggest         = "prices-to-suggest"
	FlagPlugin                  = "plugin"
)

// Config defines the configuration of the rosetta server
type Config struct {
	// Blockchain defines the blockchain name
	// defaults to DefaultBlockchain
	Blockchain string
	// Network defines the network name
	Network string
	// TendermintRPC defines the endpoint to connect to
	// CometBFT RPC, specifying 'tcp://' before is not
	// required, usually it's at port 26657 of the
	TendermintRPC string
	// GRPCEndpoint defines the cosmos application gRPC endpoint
	// usually it is located at 9090 port
	GRPCEndpoint string
	// Addr defines the default address to bind the rosetta server to
	// defaults to DefaultAddr
	Addr string
	// Retries defines the maximum number of retries
	// rosetta will do before quitting
	Retries int
	// Offline defines if the server must be run in offline mode
	Offline bool
	// EnableFeeSuggestion indicates to use fee suggestion when `construction/metadata` is called without gas limit and price
	EnableFeeSuggestion bool
	// GasToSuggest defines the gas limit for fee suggestion
	GasToSuggest int
	// DenomToSuggest defines the default denom for fee suggestion
	DenomToSuggest string
	// GasPrices defines the gas prices for fee suggestion
	GasPrices sdk.DecCoins
	// Codec overrides the default data and construction api client codecs
	Codec *codec.ProtoCodec
	// InterfaceRegistry overrides the default data and construction api interface registry
	InterfaceRegistry codectypes.InterfaceRegistry
}

// NetworkIdentifier returns the network identifier given the configuration
func (c *Config) NetworkIdentifier() *types.NetworkIdentifier {
	return &types.NetworkIdentifier{
		Blockchain: c.Blockchain,
		Network:    c.Network,
	}
}

// validate validates a configuration and sets
// its defaults in case they were not provided
func (c *Config) validate() error {
	if (c.Codec == nil) != (c.InterfaceRegistry == nil) {
		return crgerrs.WrapError(crgerrs.ErrConfig, "codec and interface registry must be both different from nil or nil")
	}

	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
	if c.Blockchain == "" {
		c.Blockchain = DefaultBlockchain
	}
	if c.Retries == 0 {
		c.Retries = DefaultRetries
	}
	// these are must
	if c.Network == "" {
		return crgerrs.WrapError(crgerrs.ErrConfig, "network not provided")
	}
	if c.GasToSuggest <= 0 {
		return crgerrs.WrapError(crgerrs.ErrConfig, "gas to suggest must be positive")
	}
	if c.EnableFeeSuggestion {
		found := false
		for i := 0; i < c.GasPrices.Len(); i++ {
			if c.GasPrices.GetDenomByIndex(i) == c.DenomToSuggest {
				found = true
				break
			}
		}
		if !found {
			return crgerrs.WrapError(crgerrs.ErrConfig, "default suggest denom is not found in prices to suggest")
		}
	}

	// these are optional but it must be online
	if c.GRPCEndpoint == "" {
		return crgerrs.WrapError(crgerrs.ErrConfig, "grpc endpoint not provided")
	}
	if c.TendermintRPC == "" {
		return crgerrs.WrapError(crgerrs.ErrConfig, "cometbft rpc not provided")
	}
	validatedURL, err := c.validateURL(c.TendermintRPC)
	if err != nil {
		return err
	}
	c.TendermintRPC = validatedURL

	return nil
}

func (c *Config) validateURL(tendermintRPC string) (string, error) {
	u, err := url.Parse(tendermintRPC)
	if err != nil {
		return "", crgerrs.WrapError(crgerrs.ErrConfig, err.Error())
	}

	if u.Port() == HTTPSPORT && u.Scheme != HTTPS {
		u.Scheme = HTTPS
	}

	if u.Port() == "" {
		switch u.Scheme {
		case HTTP, TCP:
			u.Host += ":80"
		case HTTPS:
			u.Host += ":443"
		}
	}

	return u.String(), nil
}

// WithCodec extends the configuration with a predefined Codec
func (c *Config) WithCodec(ir codectypes.InterfaceRegistry, cdc *codec.ProtoCodec) {
	c.Codec = cdc
	c.InterfaceRegistry = ir
}

// FromFlags gets the configuration from flags
func FromFlags(flags *pflag.FlagSet) (*Config, error) {
	blockchain, err := flags.GetString(FlagBlockchain)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting blockchain flag %s", err.Error()))
	}
	network, err := flags.GetString(FlagNetwork)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting network flag %s", err.Error()))
	}
	tendermintRPC, err := flags.GetString(FlagTendermintEndpoint)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting tendermintRPC flag %s", err.Error()))
	}
	gRPCEndpoint, err := flags.GetString(FlagGRPCEndpoint)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting gRPCEndpoint flag %s", err.Error()))
	}
	addr, err := flags.GetString(FlagAddr)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting addr flag %s", err.Error()))
	}
	retries, err := flags.GetInt(FlagRetries)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting retries flag %s", err.Error()))
	}
	offline, err := flags.GetBool(FlagOffline)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting offline flag %s", err.Error()))
	}
	enableDefaultFeeSuggestion, err := flags.GetBool(FlagEnableFeeSuggestion)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting enableDefaultFeeSuggestion flag %s", err.Error()))
	}
	gasToSuggest, err := flags.GetInt(FlagGasToSuggest)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting gasToSuggest flag %s", err.Error()))
	}
	denomToSuggest, err := flags.GetString(FlagDenomToSuggest)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting denomToSuggest flag %s", err.Error()))
	}

	var prices sdk.DecCoins
	if enableDefaultFeeSuggestion {
		pricesToSuggest, err := flags.GetString(FlagPricesToSuggest)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while getting pricesToSuggest flag %s", err.Error()))
		}
		prices, err = sdk.ParseDecCoins(pricesToSuggest)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while parcing prices from the sdk %s", err.Error()))
		}
	}

	conf := &Config{
		Blockchain:          blockchain,
		Network:             network,
		TendermintRPC:       tendermintRPC,
		GRPCEndpoint:        gRPCEndpoint,
		Addr:                addr,
		Retries:             retries,
		Offline:             offline,
		EnableFeeSuggestion: enableDefaultFeeSuggestion,
		GasToSuggest:        gasToSuggest,
		DenomToSuggest:      denomToSuggest,
		GasPrices:           prices,
	}
	err = conf.validate()
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while validating configs %s", err.Error()))
	}
	return conf, nil
}

func ServerFromConfig(conf *Config) (crg.Server, error) {
	err := conf.validate()
	if err != nil {
		return crg.Server{}, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while validating configs %s", err.Error()))
	}
	client, err := NewClient(conf)
	if err != nil {
		return crg.Server{}, crgerrs.WrapError(crgerrs.ErrConfig, fmt.Sprintf("while creating a new client from configs %s", err.Error()))
	}
	return crg.NewServer(
		crg.Settings{
			Network: &types.NetworkIdentifier{
				Blockchain: conf.Blockchain,
				Network:    conf.Network,
			},
			Client:    client,
			Listen:    conf.Addr,
			Offline:   conf.Offline,
			Retries:   conf.Retries,
			RetryWait: 15 * time.Second,
		})
}

// SetFlags sets the configuration flags to the given flagset
func SetFlags(flags *pflag.FlagSet) {
	flags.String(FlagBlockchain, DefaultBlockchain, "the blockchain type")
	flags.String(FlagNetwork, DefaultNetwork, "the network name")
	flags.String(FlagTendermintEndpoint, DefaultCometEndpoint, "the CometBFT rpc endpoint, without tcp://")
	flags.String(FlagGRPCEndpoint, DefaultGRPCEndpoint, "the app gRPC endpoint")
	flags.String(FlagGRPCTypesServerEndpoint, DefaultGRPCTypesServerEndpoint, "the app gRPC Server endpoint for proto messages types and reflection")
	flags.String(FlagAddr, DefaultAddr, "the address rosetta will bind to")
	flags.Int(FlagRetries, DefaultRetries, "the number of retries that will be done before quitting")
	flags.Bool(FlagOffline, DefaultOffline, "run rosetta only with construction API")
	flags.Bool(FlagEnableFeeSuggestion, DefaultEnableFeeSuggestion, "enable default fee suggestion")
	flags.Int(FlagGasToSuggest, clientflags.DefaultGasLimit, "default gas for fee suggestion")
	flags.String(FlagDenomToSuggest, DenomToSuggest, "default denom for fee suggestion")
	flags.String(FlagPricesToSuggest, DefaultPrices, "default prices for fee suggestion")
	flags.String(FlagPlugin, DefaultPlugin, "plugin folder name")
}
