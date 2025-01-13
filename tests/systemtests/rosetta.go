package systemtests

import (
	"bufio"
	"container/ring"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"cosmossdk.io/systemtests"
)

type cleanUpFn func()

var rosetta rosettaRunner

type rosettaRunner struct {
	execBinary      string
	Addr            string // the address rosetta will bind to (default ":8080")
	Blockchain      string // the blockchain type (default "app")
	DenomToSuggest  string // default denom for fee suggestion (default "uatom")
	GRPC            string // the app gRPC endpoint (default "localhost:9090")
	GRPCTypesServer string // the app gRPC Server endpoint for proto messages types and reflection
	Network         string // the network name (default "network")
	Plugin          string // plugin folder name
	Tendermint      string // CometBFT rpc endpoint
	Offline         bool   // run rosetta only with construction API
	verbose         bool
	out             io.Writer
	outBuff         *ring.Ring
	errBuff         *ring.Ring
	cleanupFn       []cleanUpFn

	pid       int
	outputDir string
}

func newRosettaRunner(binary, denom, grpcTypesServer, plugin string, offline, verbose bool) rosettaRunner {
	execBinary := filepath.Join(systemtests.WorkDir, "binaries", binary)
	return rosettaRunner{
		execBinary:      execBinary,
		Addr:            "localhost:8080",
		Blockchain:      "testing",
		DenomToSuggest:  denom,
		GRPC:            "localhost:9090",
		GRPCTypesServer: grpcTypesServer,
		Network:         "cosmos",
		Plugin:          plugin,
		Tendermint:      "tcp://localhost:26657",
		Offline:         offline,
		out:             os.Stdout,
		outBuff:         ring.New(100),
		errBuff:         ring.New(100),
		verbose:         verbose,
		pid:             -1,
		outputDir:       "./testnet",
	}
}

func (r *rosettaRunner) start(t *testing.T) {
	args := []string{
		"--blockchain", r.Blockchain,
		"--network", r.Network,
		"--tendermint", r.Tendermint,
		"--addr", r.Addr,
		"--grpc", r.GRPC,
		"--grpc-types-server", r.GRPCTypesServer,
		"--plugin", r.Plugin,
	}

	r.log("Start Rosetta\n")
	r.logf("Execute `%s %s`\n", r.execBinary, strings.Join(args, " "))
	cmd := exec.Command(locateExecutable(r.execBinary), args...)
	cmd.Dir = r.outputDir
	r.watchLogs(cmd)

	require.NoError(t, cmd.Start())

	r.pid = cmd.Process.Pid

	r.awaitRosettaUp(t)
}

func (r *rosettaRunner) awaitRosettaUp(t *testing.T) {
	t.Log("Waiting for rosetta to start\n")

	client := resty.New().SetHostURL("http://" + r.Addr)
	for i := 0; i < 10; i++ {
		res, err := client.R().SetHeader("Content-Type", "application/json; charset=UTF-8").
			SetBody("{}").
			Post("/network/list")
		if err == nil {
			bk := gjson.GetBytes(res.Body(), "network_identifiers.#.blockchain").Array()[0].String()
			require.Equal(t, bk, "testing")
			t.Log("Rosetta has been started\n")
			return
		}
		time.Sleep(time.Second * 2)
	}
	t.Fatalf("failed to connect to Rosetta")
}

func (r *rosettaRunner) restart(t *testing.T) {
	t.Log("Restarting Rosetta\n")
	assert.NoError(t, r.stop())

	r.start(t)
}

func (r *rosettaRunner) stop() error {
	if r.pid == -1 {
		return nil
	}

	p, err := os.FindProcess(r.pid)
	if err != nil {
		return err
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		r.logf("failed to stop node with pid %d: %s\n", p.Pid, err)
	}
	time.Sleep(time.Second * 2)
	return nil
}

func (r *rosettaRunner) log(msg string) {
	if r.verbose {
		_, _ = fmt.Fprint(r.out, msg)
	}
}

func (r *rosettaRunner) logf(msg string, args ...interface{}) {
	r.log(fmt.Sprintf(msg, args...))
}

// watchLogs stores stdout/stderr in a file and in a ring buffer to output the last n lines on test error
func (r *rosettaRunner) watchLogs(cmd *exec.Cmd) {
	logfile, err := os.Create(filepath.Join(systemtests.WorkDir, r.outputDir, "rosetta.out"))
	if err != nil {
		panic(fmt.Sprintf("open logfile error %#+v", err))
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("stderr reader error %#+v", err))
	}
	stopRingBuffer := make(chan struct{})
	go appendToBuf(io.TeeReader(errReader, logfile), r.errBuff, stopRingBuffer)

	outReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(fmt.Sprintf("stdout reader error %#+v", err))
	}
	go appendToBuf(io.TeeReader(outReader, logfile), r.outBuff, stopRingBuffer)
	r.cleanupFn = append(r.cleanupFn, func() {
		close(stopRingBuffer)
		_ = logfile.Close()
	})
}

func appendToBuf(r io.Reader, b *ring.Ring, stop <-chan struct{}) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-stop:
			return
		default:
		}
		text := scanner.Text()
		b.Value = text
		b = b.Next()
	}
}

// locateExecutable looks up the binary on the OS path.
func locateExecutable(file string) string {
	if strings.TrimSpace(file) == "" {
		panic("executable binary name must not be empty")
	}
	path, err := exec.LookPath(file)
	if err != nil {
		panic(fmt.Sprintf("unexpected error with file %q: %s", file, err.Error()))
	}
	if path == "" {
		panic(fmt.Sprintf("%q not found", file))
	}
	return path
}
