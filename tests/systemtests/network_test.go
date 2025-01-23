//go:build system_test

package rossettaSystemTests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"
)

func TestNetwork(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	// network/list endpoint
	res, err := rosettaRest.networkList()
	assert.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(res, "network_identifiers.0.blockchain").String(), rosetta.Blockchain)
	assert.Equal(t, gjson.GetBytes(res, "network_identifiers.0.network").String(), rosetta.Network)

	// network/status endpoint
	res, err = rosettaRest.networkStatus()
	assert.NoError(t, err)
	assert.Greater(t, gjson.GetBytes(res, "current_block_identifier.index").Int(), int64(1))
	assert.Equal(t, gjson.GetBytes(res, "genesis_block_identifier.index").Int(), int64(1))
	assert.Equal(t, gjson.GetBytes(res, "sync_status.stage").String(), "synced")

	// network/options
	res, err = rosettaRest.networkOptions()
	assert.NoError(t, err)
	assert.Equal(t, len(gjson.GetBytes(res, "allow.operation_statuses").Array()), 2)
}
