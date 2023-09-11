package rosetta

import (
	"fmt"
	"os"
	"plugin"

	crgerrs "github.com/cosmos/rosetta/lib/errors"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func LoadPlugin(ir codectypes.InterfaceRegistry, pluginLocation string) (err error) {
	pluginPathMain := fmt.Sprintf("./plugins/%s/main.so", pluginLocation)

	if _, err := os.Stat(pluginPathMain); os.IsExist(err) {
		return crgerrs.WrapError(crgerrs.ErrPlugin, fmt.Sprintf("Plugin file '%s' does not exist %s", pluginPathMain, err.Error()))
	}

	// load module
	plug, err := plugin.Open(pluginPathMain)
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrPlugin, fmt.Sprintf("There was an error while opening plugin on %s - %s", pluginPathMain, err.Error()))
	}

	initZone, err := plug.Lookup("InitZone")
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrPlugin, fmt.Sprintf("There was an error while initializing the zone %s", err.Error()))
	}
	initZone.(func())()

	registerInterfaces, err := plug.Lookup("RegisterInterfaces")
	if err != nil {
		return crgerrs.WrapError(crgerrs.ErrPlugin, fmt.Sprintf("There was an error while registering interfaces %s", err.Error()))
	}

	registerInterfaces.(func(codectypes.InterfaceRegistry))(ir)
	return nil
}
