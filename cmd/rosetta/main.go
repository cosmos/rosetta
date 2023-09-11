package main

import (
	"os"

	"github.com/cosmos/rosetta"

	"cosmossdk.io/log"
	rosettaCmd "github.com/cosmos/rosetta/cmd"
)

func main() {
	var (
		cdc, interfaceRegistry = rosetta.MakeCodec()
		logger                 = log.NewLogger(os.Stdout).With(log.ModuleKey, "rosetta")
	)

	if err := rosettaCmd.RosettaCommand(interfaceRegistry, cdc).Execute(); err != nil {
		logger.Error("failed to run rosetta", "error", err)
		os.Exit(1)
	}
}
