package main

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"

	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func InitZone() {
	config := sdk.GetConfig()
	prefix := "cosmos-hub"
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
