package service

import (
	"context"

	"cosmossdk.io/rosetta/lib/errors"
	coinbase "github.com/coinbase/rosetta-sdk-go/types"
)

func (on OnlineNetwork) NetworkList(_ context.Context, _ *coinbase.MetadataRequest) (*coinbase.NetworkListResponse, *coinbase.Error) {
	return &coinbase.NetworkListResponse{NetworkIdentifiers: []*coinbase.NetworkIdentifier{on.network}}, nil
}

func (on OnlineNetwork) NetworkOptions(_ context.Context, _ *coinbase.NetworkRequest) (*coinbase.NetworkOptionsResponse, *coinbase.Error) {
	return on.networkOptions, nil
}

func (on OnlineNetwork) NetworkStatus(ctx context.Context, _ *coinbase.NetworkRequest) (*coinbase.NetworkStatusResponse, *coinbase.Error) {
	syncStatus, err := on.client.Status(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	block, err := on.client.BlockByHeight(ctx, syncStatus.CurrentIndex)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	oldestBlockIdentifier, err := on.client.OldestBlock(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	genesisBlock, err := on.client.GenesisBlock(ctx)
	if err != nil {
		genesisBlock, err = on.client.InitialHeightBlock(ctx)
		if err != nil {
			genesisBlock = oldestBlockIdentifier
		}
	}

	peers, err := on.client.Peers(ctx)
	if err != nil {
		return nil, errors.ToRosetta(err)
	}

	return &coinbase.NetworkStatusResponse{
		CurrentBlockIdentifier: block.Block,
		CurrentBlockTimestamp:  block.MillisecondTimestamp,
		GenesisBlockIdentifier: genesisBlock.Block,
		OldestBlockIdentifier:  oldestBlockIdentifier.Block,
		SyncStatus:             syncStatus,
		Peers:                  peers,
	}, nil
}
