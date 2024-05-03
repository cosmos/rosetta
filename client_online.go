package rosetta

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/version"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/cometbft/cometbft/rpc/client/http"
	"google.golang.org/grpc"

	crgerrs "github.com/cosmos/rosetta/lib/errors"
	crgtypes "github.com/cosmos/rosetta/lib/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"

	tmrpc "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// interface assertion
var _ crgtypes.Client = (*Client)(nil)

const (
	defaultNodeTimeout = time.Minute
	tmWebsocketPath    = "/websocket"
)

// Client implements a single network client to interact with cosmos based chains
type Client struct {
	supportedOperations []string

	config *Config

	auth  auth.QueryClient
	bank  bank.QueryClient
	tmRPC tmrpc.Client

	version string

	converter Converter
}

// NewClient instantiates a new online servicer
func NewClient(cfg *Config) (*Client, error) {
	info := version.NewInfo()

	v := info.Version
	if v == "" {
		v = "unknown"
	}

	txConfig := authtx.NewTxConfig(cfg.Codec, authtx.DefaultSignModes)

	var supportedOperations []string
	for _, ii := range cfg.InterfaceRegistry.ListImplementations(sdk.MsgInterfaceProtoName) {
		_, err := cfg.InterfaceRegistry.Resolve(ii)
		if err != nil {
			continue
		}

		supportedOperations = append(supportedOperations, ii)
	}

	supportedOperations = append(
		supportedOperations,
		bank.EventTypeCoinSpent,
		bank.EventTypeCoinReceived,
		bank.EventTypeCoinBurn,
	)

	return &Client{
		supportedOperations: supportedOperations,
		config:              cfg,
		auth:                nil,
		bank:                nil,
		tmRPC:               nil,
		version:             fmt.Sprintf("%s/%s", info.AppName, v),
		converter:           NewConverter(cfg.Codec, cfg.InterfaceRegistry, txConfig),
	}, nil
}

// ---------- cosmos-rosetta-gateway.types.Client implementation ------------ //

// Bootstrap is gonna connect the client to the endpoints
func (c *Client) Bootstrap() error {
	grpcConn, err := grpc.Dial(c.config.GRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("dialing grpc endpoint %s", err.Error()))
	}

	tmRPC, err := http.New(c.config.TendermintRPC, tmWebsocketPath)
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rpc path %s", err.Error()))
	}

	authClient := auth.NewQueryClient(grpcConn)
	bankClient := bank.NewQueryClient(grpcConn)

	c.auth = authClient
	c.bank = bankClient
	c.tmRPC = tmRPC

	return nil
}

// Ready performs a health check and returns an error if the client is not ready.
func (c *Client) Ready() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultNodeTimeout)
	defer cancel()
	_, err := c.tmRPC.Health(ctx)
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting grpc health %s", err.Error()))
	}

	_, err = c.tmRPC.Status(ctx)
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting grpc status %s", err.Error()))
	}

	_, err = c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting bank total supply %s", err.Error()))
	}
	return nil
}

func (c *Client) GenesisBlock(ctx context.Context) (crgtypes.BlockResponse, error) {
	var genesisHeight int64 = 1
	return c.BlockByHeight(ctx, &genesisHeight)
}

func (c *Client) InitialHeightBlock(ctx context.Context) (crgtypes.BlockResponse, error) {
	genesisChunk, err := c.tmRPC.GenesisChunked(ctx, 0)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting bank total supply %s", err.Error()))
	}
	heightNum, err := extractInitialHeightFromGenesisChunk(genesisChunk.Data)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting height from genesis chunk %s", err.Error()))
	}
	return c.BlockByHeight(ctx, &heightNum)
}

func (c *Client) OldestBlock(ctx context.Context) (crgtypes.BlockResponse, error) {
	status, err := c.tmRPC.Status(ctx)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting oldest block %s", err.Error()))
	}
	return c.BlockByHeight(ctx, &status.SyncInfo.EarliestBlockHeight)
}

func (c *Client) accountInfo(ctx context.Context, addr string, height *int64) (*SignerData, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}

	accountInfo, err := c.auth.Account(ctx, &auth.QueryAccountRequest{
		Address: addr,
	})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}

	signerData, err := c.converter.ToRosetta().SignerData(accountInfo.Account)
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}
	return signerData, nil
}

func (c *Client) Balances(ctx context.Context, addr string, height *int64) ([]*rosettatypes.Amount, error) {
	if height != nil {
		strHeight := strconv.FormatInt(*height, 10)
		ctx = metadata.AppendToOutgoingContext(ctx, grpctypes.GRPCBlockHeightHeader, strHeight)
	}

	balance, err := c.bank.AllBalances(ctx, &bank.QueryAllBalancesRequest{
		Address: addr,
	})
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}

	availableCoins, err := c.coins(ctx)
	if err != nil {
		return nil, crgerrs.FromGRPCToRosettaError(err)
	}

	return c.converter.ToRosetta().Amounts(balance.Balances, availableCoins), nil
}

func (c *Client) BlockByHash(ctx context.Context, hash string) (crgtypes.BlockResponse, error) {
	bHash, err := hex.DecodeString(hash)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("invalid block hash %s", err.Error()))
	}

	block, err := c.tmRPC.BlockByHash(ctx, bHash)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block by hash %s", err.Error()))
	}

	return c.converter.ToRosetta().BlockResponse(block), nil
}

func (c *Client) BlockByHeight(ctx context.Context, height *int64) (crgtypes.BlockResponse, error) {
	block, err := c.tmRPC.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block by height %s", err.Error()))
	}

	return c.converter.ToRosetta().BlockResponse(block), nil
}

func (c *Client) BlockTransactionsByHash(ctx context.Context, hash string) (crgtypes.BlockTransactionsResponse, error) {
	// TODO(fdymylja): use a faster path, by searching the block by hash, instead of doing a double query operation
	blockResp, err := c.BlockByHash(ctx, hash)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block transactions by hash %s", err.Error()))
	}

	return c.blockTxs(ctx, &blockResp.Block.Index)
}

func (c *Client) BlockTransactionsByHeight(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	blockTxResp, err := c.blockTxs(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block transactions by height %s", err.Error()))
	}
	return blockTxResp, nil
}

// Coins f etches the existing coins in the application
func (c *Client) coins(ctx context.Context) (sdk.Coins, error) {
	var result sdk.Coins

	supply, err := c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{})
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting coins supply %s", err.Error()))
	}

	pages := supply.GetPagination().GetTotal()
	for i := uint64(0); i < pages; i++ {
		// get next key
		page := supply.GetPagination()
		if page == nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting supply pagination %s", err.Error()))
		}
		nextKey := page.GetNextKey()

		supply, err = c.bank.TotalSupply(ctx, &bank.QueryTotalSupplyRequest{Pagination: &query.PageRequest{Key: nextKey}})
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting supply from bank %s", err.Error()))
		}

		result = append(result[:0], supply.Supply[:]...)
	}

	return result, nil
}

func (c *Client) TxOperationsAndSignersAccountIdentifiers(signed bool, txBytes []byte) (ops []*rosettatypes.Operation, signers []*rosettatypes.AccountIdentifier, err error) {
	switch signed {
	case false:
		rosTx, err := c.converter.ToRosetta().Tx(txBytes, nil)
		if err != nil {
			return nil, nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rosseta Tx format %s", err.Error()))
		}
		return rosTx.Operations, nil, nil
	default:
		ops, signers, err = c.converter.ToRosetta().OpsAndSigners(txBytes)
		return
	}
}

// GetTx returns a transaction given its hash. For Rosetta we  make a synthetic transaction for BeginBlock
//
//	and EndBlock to adhere to balance tracking rules.
func (c *Client) GetTx(ctx context.Context, hash string) (*rosettatypes.Transaction, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("bad tx hash %s", err.Error()))
	}

	// get tx type and hash
	txType, hashBytes := c.converter.ToSDK().HashToTxType(hashBytes)

	// construct rosetta tx
	switch txType {
	// handle begin block hash
	// handle deliver tx hash
	case DeliverTxTx:
		rawTx, err := c.tmRPC.Tx(ctx, hashBytes, true)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting tx %s", err.Error()))
		}
		return c.converter.ToRosetta().Tx(rawTx.Tx, &rawTx.TxResult)
	// handle end block hash
	case FinalizeBlockTx:
		// get block height by hash
		block, err := c.tmRPC.BlockByHash(ctx, hashBytes)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block by hash %s", err.Error()))
		}

		// get block txs
		fullBlock, err := c.blockTxs(ctx, &block.Block.Height)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting block by hash %s", err.Error()))
		}

		// get last tx
		return fullBlock.Transactions[len(fullBlock.Transactions)-1], nil
	// unrecognized tx
	default:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("invalid tx hash provided: %s", hash))
	}
}

// GetUnconfirmedTx gets an unconfirmed transaction given its hash
func (c *Client) GetUnconfirmedTx(ctx context.Context, hash string) (*rosettatypes.Transaction, error) {
	res, err := c.tmRPC.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrNotFound, fmt.Sprintf("unconfirmed tx not found %s", err.Error()))
	}

	hashAsBytes, err := hex.DecodeString(hash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, fmt.Sprintf("invalid hash %s", err.Error()))
	}

	// assert that correct tx length is provided
	switch len(hashAsBytes) {
	default:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("unrecognized tx size: %d", len(hashAsBytes)))
	case FinalizeBlockTxSize:
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "endblock and begin block txs cannot be unconfirmed")
	case DeliverTxSize:
		break
	}

	// iterate over unconfirmed txs to find the one with matching hash
	for _, unconfirmedTx := range res.Txs {
		if !bytes.Equal(unconfirmedTx.Hash(), hashAsBytes) {
			continue
		}

		return c.converter.ToRosetta().Tx(unconfirmedTx, nil)
	}
	return nil, crgerrs.WrapError(crgerrs.ErrNotFound, "transaction not found in mempool: "+hash)
}

// Mempool returns the unconfirmed transactions in the mempool
func (c *Client) Mempool(ctx context.Context) ([]*rosettatypes.TransactionIdentifier, error) {
	txs, err := c.tmRPC.UnconfirmedTxs(ctx, nil)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting unconfirmed tx %s", err.Error()))
	}

	return c.converter.ToRosetta().TxIdentifiers(txs.Txs), nil
}

// Peers gets the number of peers
func (c *Client) Peers(ctx context.Context) ([]*rosettatypes.Peer, error) {
	netInfo, err := c.tmRPC.NetInfo(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, "getting network information "+err.Error())
	}
	return c.converter.ToRosetta().Peers(netInfo.Peers), nil
}

func (c *Client) Status(ctx context.Context) (*rosettatypes.SyncStatus, error) {
	status, err := c.tmRPC.Status(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting network information %s", err.Error()))
	}
	return c.converter.ToRosetta().SyncStatus(status), err
}

func (c *Client) PostTx(txBytes []byte) (*rosettatypes.TransactionIdentifier, map[string]interface{}, error) {
	// sync ensures it will go through checkTx
	res, err := c.tmRPC.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		return nil, nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting BroadcastTxSync %s", err.Error()))
	}
	// check if tx was broadcast successfully
	if res.Code != abcitypes.CodeTypeOK {
		return nil, nil, crgerrs.WrapError(
			crgerrs.ErrUnknown,
			fmt.Sprintf("transaction broadcast failure: (%d) %s ", res.Code, res.Log),
		)
	}

	return &rosettatypes.TransactionIdentifier{
			Hash: fmt.Sprintf("%X", res.Hash),
		},
		map[string]interface{}{
			Log: res.Log,
		}, nil
}

// construction endpoints

// ConstructionMetadataFromOptions builds the metadata given the options
func (c *Client) ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error) {
	if len(options) == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "options length is 0")
	}

	constructionOptions := new(PreprocessOperationsOptionsResponse)

	err = constructionOptions.FromMetadata(options)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, fmt.Sprintf("getting metadata %s", err.Error()))
	}

	// if default fees suggestion is enabled and gas limit or price is unset, use default
	if c.config.EnableFeeSuggestion {
		if constructionOptions.GasLimit <= 0 {
			constructionOptions.GasLimit = uint64(c.config.GasToSuggest)
		}
		if constructionOptions.GasPrice == "" {
			denom := c.config.DenomToSuggest
			constructionOptions.GasPrice = c.config.GasPrices.AmountOf(denom).String() + denom
		}
	}

	if constructionOptions.GasLimit > 0 && constructionOptions.GasPrice != "" {
		gasPrice, err := sdk.ParseDecCoin(constructionOptions.GasPrice)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("parsing gas price %s", err.Error()))
		}
		if !gasPrice.IsPositive() {
			return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "gas price must be positive")
		}
	}

	signersData := make([]*SignerData, len(constructionOptions.ExpectedSigners))

	for i, signer := range constructionOptions.ExpectedSigners {
		accountInfo, err := c.accountInfo(ctx, signer, nil)
		if err != nil {
			return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting account info %s", err.Error()))
		}

		signersData[i] = accountInfo
	}

	status, err := c.tmRPC.Status(ctx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rpc status %s", err.Error()))
	}

	metadataResp := ConstructionMetadata{
		ChainID:     status.NodeInfo.Network,
		SignersData: signersData,
		GasLimit:    constructionOptions.GasLimit,
		GasPrice:    constructionOptions.GasPrice,
		Memo:        constructionOptions.Memo,
	}

	return metadataResp.ToMetadata()
}

func (c *Client) blockTxs(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	// get block info
	blockInfo, err := c.tmRPC.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rpc block %s", err.Error()))
	}
	// get block events
	blockResults, err := c.tmRPC.BlockResults(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rpc block results %s", err.Error()))
	}

	if len(blockResults.TxsResults) != len(blockInfo.Block.Txs) {
		return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, "block results transactions do now match block transactions")
	}
	// process begin and end block txs
	FinalizeBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: c.converter.ToRosetta().FinalizeBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: AddOperationIndexes(
			nil,
			c.converter.ToRosetta().BalanceOps(StatusTxSuccess, blockResults.FinalizeBlockEvents),
		),
	}

	deliverTx := make([]*rosettatypes.Transaction, len(blockInfo.Block.Txs))
	// process normal txs
	for i, tx := range blockInfo.Block.Txs {
		rosTx, err := c.converter.ToRosetta().Tx(tx, blockResults.TxsResults[i])
		if err != nil {
			return crgtypes.BlockTransactionsResponse{}, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting rosetta tx %s", err.Error()))
		}
		deliverTx[i] = rosTx
	}

	finalTxs := make([]*rosettatypes.Transaction, 0, 1+len(deliverTx))
	finalTxs = append(finalTxs, deliverTx...)
	finalTxs = append(finalTxs, FinalizeBlockTx)

	return crgtypes.BlockTransactionsResponse{
		BlockResponse: c.converter.ToRosetta().BlockResponse(blockInfo),
		Transactions:  finalTxs,
	}, nil
}

var initialHeightRE = regexp.MustCompile(`"initial_height":"(\d+)"`)

func extractInitialHeightFromGenesisChunk(genesisChunk string) (int64, error) {
	firstChunk, err := base64.StdEncoding.DecodeString(genesisChunk)
	if err != nil {
		return 0, crgerrs.WrapError(crgerrs.ErrOnlineClient, fmt.Sprintf("getting first chunk %s", err.Error()))
	}

	matches := initialHeightRE.FindStringSubmatch(string(firstChunk))
	if len(matches) != 2 {
		return 0, crgerrs.WrapError(crgerrs.ErrOnlineClient, "failed to fetch initial_height")
	}

	heightStr := matches[1]
	return strconv.ParseInt(heightStr, 10, 64)
}
