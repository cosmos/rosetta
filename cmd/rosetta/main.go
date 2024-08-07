package main

import (
	"os"

	"cosmossdk.io/log"

	"github.com/cosmos/rosetta"
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
