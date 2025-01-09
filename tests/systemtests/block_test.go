//go:build system_test

package systemtests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
	"testing"

	"cosmossdk.io/systemtests"
)

func TestBlock(t *testing.T) {
	sut.ResetChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, verbose)

	// add genesis account with some tokens
	fromAddr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", fromAddr, "10000000stake"},
	)
	toAddr := cli.AddKey("account2")

	sut.StartChain(t)
	//sut.AwaitNBlocks(t, 1)
	rosetta.restart(t)

	// stake tokens
	rsp := cli.RunAndWait("tx", "bank", "send", fromAddr, toAddr, "1000000stake")
	systemtests.RequireTxSuccess(t, rsp)

	resClient := newRestClient(rosetta.Addr, rosetta.Network, rosetta.Blockchain)
	height := gjson.Get(rsp, "height").String()
	res, err := resClient.block(height)
	assert.NoError(t, err)
	fmt.Println(res)
}
