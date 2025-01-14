//go:build system_test

package systemtests

import (
	"github.com/tidwall/gjson"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMempool(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	res, err := rosettaRest.mempool()
	assert.NoError(t, err)
	assert.NotNil(t, gjson.GetBytes(res, "transaction_identifiers").Array())
}
