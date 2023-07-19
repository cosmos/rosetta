package rosetta

import (
	"fmt"
	"os"
	"plugin"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func LoadPlugin(ir codectypes.InterfaceRegistry, pluginLocation string) (err error) {
	pluginPathMain := fmt.Sprintf("./plugins/%s/main.so", pluginLocation)

	if _, err := os.Stat(pluginPathMain); os.IsExist(err) {
		fmt.Printf("Plugin file '%s' does not exist...\n", pluginPathMain)
		return err
	}

	// load module
	plug, err := plugin.Open(pluginPathMain)
	if err != nil {
		fmt.Println("There was an error while opening the plugin...", err)
		return err
	}

	initZone, err := plug.Lookup("InitZone")
	if err != nil {
		fmt.Println("There was an error while initializing the zone.", err)
		return err
	}
	initZone.(func())()

	registerInterfaces, err := plug.Lookup("RegisterInterfaces")
	if err != nil {
		fmt.Println("There was an error while registering interfaces...", err)
		return err
	}

	registerInterfaces.(func(codectypes.InterfaceRegistry))(ir)
	return err
}
