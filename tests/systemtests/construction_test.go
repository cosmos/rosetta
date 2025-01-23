//go:build system_test

package rossetaSystemTests

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func TestDerive(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	pubKey := secp256k1.GenPrivKey().PubKey()
	hexPk := strings.Split(pubKey.String(), "{")[1]

	res, err := rosettaRest.constructionDerive(hexPk[:len(hexPk)-1])
	assert.NoError(t, err)

	addr, err := address.NewBech32Codec("cosmos").BytesToString(pubKey.Address().Bytes())
	assert.NoError(t, err)
	assert.Equal(t, addr, gjson.GetBytes(res, "address").String())
}

func TestHash(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := systemtests.NewCLIWrapper(t, sut, verbose)
	fromAddr := cli.AddKey("account1")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", fromAddr, "10000000stake"},
	)
	toAddr := cli.AddKey("account2")

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	rsp := cli.RunCommandWithArgs(cli.WithTXFlags("tx", "bank", "send", fromAddr, toAddr, "10stake", "--generate-only")...)
	tempFile := systemtests.StoreTempFile(t, []byte(rsp))
	rsp = cli.RunCommandWithArgs(cli.WithTXFlags("tx", "sign", tempFile.Name(), "--from", fromAddr)...)
	tempFile = systemtests.StoreTempFile(t, []byte(rsp))
	rsp = cli.RunCommandWithArgs("tx", "encode", tempFile.Name())

	txBytes, err := base64.StdEncoding.DecodeString(rsp)
	assert.NoError(t, err)
	hexTx := hex.EncodeToString(txBytes)

	res, err := rosettaRest.constructionHash(hexTx)
	assert.NoError(t, err)
	assert.NotEmpty(t, gjson.GetBytes(res, "transaction_identifier.hash"))
}

func TestMetadata(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	pubKey := secp256k1.GenPrivKey().PubKey()
	hexPk := strings.Split(pubKey.String(), "{")[1]

	metadata := make(map[string]interface{})
	metadata["gas_price"] = `"123uatom"`
	metadata["gas_limit"] = 423

	res, err := rosettaRest.constructionMetadata(hexPk, metadata)
	assert.NoError(t, err)
	assert.Equal(t, gjson.GetBytes(res, "metadata.gas_price").String(), "123uatom")
	assert.Greater(t, gjson.GetBytes(res, "suggested_fee.0.value").Int(), int64(0))
}

func TestPayloads(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	rosetta.restart(t)
	rosettaRest := newRestClient(rosetta)

	cli := systemtests.NewCLIWrapper(t, sut, verbose)
	addr := cli.GetKeyAddr("node0")
	bz, err := base64.StdEncoding.DecodeString(cli.GetPubKeyByCustomField(addr, "address"))
	assert.NoError(t, err)

	pk := secp256k1.PubKey{Key: bz}
	hexPk := strings.Split(pk.String(), "{")[1]
	hexPk = hexPk[:len(hexPk)-1]

	op := operation{
		msgType:  "/cosmos.bank.v1beta1.MsgSend",
		metadata: fmt.Sprintf(`{"from_address": "%s", "to_address": "%s", "amount":[{"amount":"123", "denom":"stake"}]}`, addr, cli.AddKey("to_address")),
	}
	res, err := rosettaRest.constructionPayloads(`"signer_data":[{"account_number":1, "sequence": 0}]`, hexPk, op)
	assert.NoError(t, err)
	assert.NotEmpty(t, gjson.GetBytes(res, "unsigned_transaction"))
	assert.Equal(t, gjson.GetBytes(res, "payloads.0.address").String(), addr)
}
