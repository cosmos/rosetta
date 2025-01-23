//go:build system_test

package rossettaSystemTests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"cosmossdk.io/systemtests"
)

func TestBlockAndBlockTransaction(t *testing.T) {
	sut.ResetChain(t)
	cli := systemtests.NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	fromAddr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", fromAddr, "10000000stake"},
	)
	toAddr := cli.AddKey("account2")
	sut.StartChain(t)

	rosetta.restart(t)

	// stake tokens
	rsp := cli.RunAndWait("tx", "bank", "send", fromAddr, toAddr, "10stake")
	systemtests.RequireTxSuccess(t, rsp)

	rosettaRest := newRestClient(rosetta)

	// test /block endpoint
	height := gjson.Get(rsp, "height").String()
	res, err := rosettaRest.block(height)
	assert.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(res, "block.block_identifier.index").String(), height)

	// test block/transaction endpoint
	blockHash := gjson.GetBytes(res, "block.block_identifier.hash").String()
	hash := gjson.GetBytes(res, "block.transactions.0.transaction_identifier.hash").String()
	res, err = rosettaRest.blockTransaction(height, blockHash, hash)
	assert.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(res, "transaction.operations.0.metadata.from_address").String(), fromAddr)
}
