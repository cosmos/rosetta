package main

import (
	bankTypes "cosmossdk.io/x/bank/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

func InitZone() {
	config := sdk.GetConfig()
	prefix := "cosmos"
	config.SetBech32PrefixForAccount(prefix, prefix+"pub")
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// ibcclienttypes.RegisterInterfaces(registry)
	// ibcLightClient.RegisterInterfaces(registry)
	sdk.RegisterInterfaces(registry)
	txtypes.RegisterInterfaces(registry)
	cryptocodec.RegisterInterfaces(registry)
	bankTypes.RegisterInterfaces(registry)
}
