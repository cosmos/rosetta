package rosetta

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"

	crgerrs "github.com/cosmos/rosetta/lib/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ---------- cosmos-rosetta-gateway.types.NetworkInformationProvider implementation ------------ //

func (c *Client) OperationStatuses() []*types.OperationStatus {
	return []*types.OperationStatus{
		{
			Status:     StatusTxSuccess,
			Successful: true,
		},
		{
			Status:     StatusTxReverted,
			Successful: false,
		},
	}
}

func (c *Client) Version() string {
	return c.version
}

func (c *Client) SupportedOperations() []string {
	return c.supportedOperations
}

// ---------- cosmos-rosetta-gateway.types.OfflineClient implementation ------------ //

func (c *Client) SignedTx(_ context.Context, txBytes []byte, signatures []*types.Signature) (signedTxBytes []byte, err error) {
	return c.converter.ToSDK().SignedTx(txBytes, signatures)
}

func (c *Client) ConstructionPayload(_ context.Context, request *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error) {
	// check if there is at least one operation
	if len(request.Operations) < 1 {
		return nil, crgerrs.WrapError(crgerrs.ErrInvalidOperation, "expected at least one operation")
	}

	tx, err := c.converter.ToSDK().UnsignedTx(request.Operations)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("getting unsigned tx %s", err.Error()))
	}

	metadata := new(ConstructionMetadata)
	if err = metadata.FromMetadata(request.Metadata); err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("getting metadata from request %s", err.Error()))
	}

	txBytes, payloads, err := c.converter.ToRosetta().SigningComponents(tx, metadata, request.PublicKeys)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("getting signed components %s", err.Error()))
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(txBytes),
		Payloads:            payloads,
	}, nil
}

func (c *Client) PreprocessOperationsToOptions(_ context.Context, req *types.ConstructionPreprocessRequest) (response *types.ConstructionPreprocessResponse, err error) {
	if len(req.Operations) == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrBadArgument, "no operations")
	}

	// now we need to parse the operations to cosmos sdk messages
	tx, err := c.converter.ToSDK().UnsignedTx(req.Operations)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("converting unsigned tx %s", err.Error()))
	}

	// get the signers
	signers, err := tx.GetSigners()
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("getting signers from unsigned tx %s", err.Error()))
	}

	signersStr := make([]string, len(signers))
	accountIdentifiers := make([]*types.AccountIdentifier, len(signers))

	for i, sig := range signers {
		addr := sdk.AccAddress(sig)
		signersStr[i] = addr.String()
		accountIdentifiers[i] = &types.AccountIdentifier{
			Address: addr.String(),
		}
	}
	// get the metadata request information
	meta := new(ConstructionPreprocessMetadata)
	err = meta.FromMetadata(req.Metadata)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("parsing metadata %s", err.Error()))
	}

	if meta.GasPrice == "" {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, "no gas price")
	}

	if meta.GasLimit == 0 {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, "no gas limit")
	}

	// prepare the options to return
	options := &PreprocessOperationsOptionsResponse{
		ExpectedSigners: signersStr,
		Memo:            meta.Memo,
		GasLimit:        meta.GasLimit,
		GasPrice:        meta.GasPrice,
	}

	metaOptions, err := options.ToMetadata()
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrOffline, fmt.Sprintf("parsing metadata %s", err.Error()))
	}
	return &types.ConstructionPreprocessResponse{
		Options:            metaOptions,
		RequiredPublicKeys: accountIdentifiers,
	}, nil
}

func (c *Client) AccountIdentifierFromPublicKey(pubKey *types.PublicKey) (*types.AccountIdentifier, error) {
	pk, err := c.converter.ToSDK().PubKey(pubKey)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrConverter, fmt.Sprintf("converting pub key to sdk %s", err.Error()))
	}

	return &types.AccountIdentifier{
		Address: sdk.AccAddress(pk.Address()).String(),
	}, nil
}
