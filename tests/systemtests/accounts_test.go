//go:build system_test

package systemtests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"cosmossdk.io/systemtests"
)

func TestAccounts(t *testing.T) {
	sut.ResetChain(t)
	cli := systemtests.NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	fromAddr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", fromAddr, "10000000stake"},
	)
	toAddr := cli.AddKey("account2")
	sut.StartChain(t)

	cli.RunAndWait("tx", "bank", "send", fromAddr, toAddr, "1000000stake")

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	// check balance after spent
	res, err := rosettaRest.accountBalance(fromAddr)
	assert.NoError(t, err)
	assert.Equal(t, int64(8999999), gjson.GetBytes(res, "balances.0.value").Int())

	// check balance at genesis, before spent
	res, err = rosettaRest.accountBalance(fromAddr, withBlockIdentifier("1"))
	assert.NoError(t, err)
	assert.Equal(t, int64(10000000), gjson.GetBytes(res, "balances.0.value").Int())
}
