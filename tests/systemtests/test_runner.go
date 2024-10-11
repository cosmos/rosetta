package systemtests

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/tests/systemtests"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	sut     *systemtests.SystemUnderTest
	verbose bool
)

func RunTests(m *testing.M) {
	waitTime := flag.Duration("wait-time", systemtests.DefaultWaitTime, "time to wait for chain events")
	nodesCount := flag.Int("nodes-count", 4, "number of nodes in the cluster")
	blockTime := flag.Duration("block-time", 1000*time.Millisecond, "block creation time")
	execBinary := flag.String("binary", "simd", "executable binary for server/ client side")
	bech32Prefix := flag.String("bech32", "cosmos", "bech32 prefix to be used with addresses")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

	// rosetta flags
	rosettaBinary := flag.String("rosetta-binary", "rosetta", "executable binary for rosetta")
	rosettaDenom := flag.String("rosetta-denom", "ustake", "rosetta denom to suggest")
	rosettaGRPCTypesServer := flag.String("rosetta-grpc-types-server", "localhost:9090", "rosetta gRPC Server endpoint for proto messages types and reflection")
	rosettaPlugin := flag.String("rosetta-plugin", "", "rosetta plugin folder name")
	rosettaOffline := flag.Bool("rosetta-offline", false, "rosetta run only with construction API")
	flag.Parse()

	requireEnoughFileHandlers(*nodesCount + 1) // +1 as tests may start another node

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	systemtests.WorkDir = dir
	if verbose {
		println("Work dir: ", systemtests.WorkDir)
	}
	initSDKConfig(*bech32Prefix)

	systemtests.DefaultWaitTime = *waitTime
	if *execBinary == "" {
		panic("executable binary name must not be empty")
	}

	sut = systemtests.NewSystemUnderTest(*execBinary, verbose, *nodesCount, *blockTime)
	sut.SetupChain() // setup chain and keyring

	rosetta = newRosettaRunner(*rosettaBinary, *rosettaDenom, *rosettaGRPCTypesServer, *rosettaPlugin, *rosettaOffline, verbose)

	// run tests
	exitCode := m.Run()

	// stop rosetta
	if err = rosetta.stop(); err != nil {
		panic(err)
	}

	// postprocess
	sut.StopChain()
	if verbose || exitCode != 0 {
		sut.PrintBuffer()
	}

	os.Exit(exitCode)
}

// requireEnoughFileHandlers uses `ulimit`
func requireEnoughFileHandlers(nodesCount int) {
	ulimit, err := exec.LookPath("ulimit")
	if err != nil || ulimit == "" { // skip when not available
		return
	}

	cmd := exec.Command(ulimit, "-n")
	cmd.Dir = systemtests.WorkDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	fileDescrCount, err := strconv.Atoi(strings.Trim(string(out), " \t\n"))
	if err != nil {
		panic(fmt.Sprintf("unexpected error :%#+v, output: %s", err, string(out)))
	}
	expFH := nodesCount * 260 // random number that worked on my box
	if fileDescrCount < expFH {
		panic(fmt.Sprintf("Fail fast. Insufficient setup. Run 'ulimit -n %d'", expFH))
	}
}

func initSDKConfig(bech32Prefix string) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(bech32Prefix, bech32Prefix+sdk.PrefixPublic)
	config.SetBech32PrefixForValidator(bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator, bech32Prefix+sdk.PrefixValidator+sdk.PrefixOperator+sdk.PrefixPublic)
	config.SetBech32PrefixForConsensusNode(bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus, bech32Prefix+sdk.PrefixValidator+sdk.PrefixConsensus+sdk.PrefixPublic)
}
