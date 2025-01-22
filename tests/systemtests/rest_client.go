package rossetaSystemTests

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type restClient struct {
	client            *resty.Client
	networkIdentifier string
}

func newRestClient(rosetta rosettaRunner) *restClient {
	client := resty.New().
		SetBaseURL("http://" + rosetta.Addr).
		SetHeaders(map[string]string{"Content-Type": "application/json", "Accept": "application/json"})
	return &restClient{
		client:            client,
		networkIdentifier: fmt.Sprintf(`"network_identifier":{"network":"%s","blockchain": "%s"}`, rosetta.Network, rosetta.Blockchain),
	}
}

func (c *restClient) networkList() ([]byte, error) {
	res, err := c.client.R().SetBody("{}").Post("/network/list")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) networkStatus() ([]byte, error) {
	res, err := c.client.R().SetBody(buildBody(c.networkIdentifier)).Post("/network/status")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) networkOptions() ([]byte, error) {
	res, err := c.client.R().SetBody(buildBody(c.networkIdentifier)).Post("/network/options")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) block(block string) ([]byte, error) {
	res, err := c.client.R().
		SetBody(fmt.Sprintf(`{%s, %s}`, c.networkIdentifier, withBlockIdentifier(block)())).
		Post("/block")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) blockTransaction(block, blockHash, txHash string) ([]byte, error) {
	res, err := c.client.R().SetBody(
		fmt.Sprintf(
			`{%s, "block_identifier":{"index":%s, "hash":"%s"}, "transaction_identifier": {"hash":"%s"}}`,
			c.networkIdentifier, block, blockHash, txHash),
	).Post("/block/transaction")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) mempool() ([]byte, error) {
	res, err := c.client.R().SetBody(
		fmt.Sprintf(`{%s}`, c.networkIdentifier),
	).Post("/mempool")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) accountBalance(address string, opts ...func() string) ([]byte, error) {
	bodyOpts := make([]string, len(opts))
	for i, opt := range opts {
		bodyOpts[i] = opt()
	}

	args := []string{c.networkIdentifier, fmt.Sprintf(`"account_identifier":{"address": "%s"}`, address)}
	args = append(args, bodyOpts...)
	jsonBody := buildBody(args...)

	res, err := c.client.R().SetBody(
		jsonBody,
	).Post("/account/balance")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func withBlockIdentifier(height string) func() string {
	return func() string {
		return fmt.Sprintf(`"block_identifier":{"index":%s}`, height)
	}
}

func buildBody(args ...string) string {
	return fmt.Sprintf(`{%s}`, strings.Join(args, ","))
}

func (c *restClient) constructionDerive(hexPk string) ([]byte, error) {
	res, err := c.client.R().SetBody(
		fmt.Sprintf(`{%s, %s}`, c.networkIdentifier, publicKeyBody(hexPk, "secp256k1")),
	).Post("/construction/derive")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func publicKeyBody(hexPk, keyType string) string {
	return fmt.Sprintf(`"public_key":{"hex_bytes":"%s", "curve_type":"%s"}`, hexPk, keyType)
}

func (c *restClient) constructionHash(hexTx string) ([]byte, error) {
	res, err := c.client.R().SetBody(
		fmt.Sprintf(`{%s, "signed_transaction": "%s"}`, c.networkIdentifier, hexTx),
	).Post("/construction/hash")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) constructionMetadata(hexPk string, options map[string]interface{}) ([]byte, error) {
	optParts := make([]string, 0, len(options))
	for k, v := range options {
		optParts = append(optParts, fmt.Sprintf(`"%s":%v`, k, v))
	}
	optionsStr := strings.Join(optParts, ",")
	body := fmt.Sprintf(`{%s, "options":{%s}, %s}`, c.networkIdentifier, optionsStr, publicKeyBody(hexPk, "secp256k1"))
	res, err := c.client.R().SetBody(
		body,
	).Post("/construction/metadata")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func (c *restClient) constructionPayloads(metadata, hexPk string, opts ...operation) ([]byte, error) {
	body := fmt.Sprintf(
		`{%s, "operations":%s, "metadata":{%s}, "public_keys":[{"hex_bytes":"%s", "curve_type":"secp256k1"}]}`,
		c.networkIdentifier, payloadsOperations(opts...), metadata, hexPk)
	res, err := c.client.R().SetBody(
		body,
	).Post("/construction/payloads")
	if err != nil {
		return nil, err
	}

	return res.Body(), nil
}

type operation struct {
	msgType  string
	metadata string
}

func (o *operation) String() string {
	return fmt.Sprintf(`"type":"%s", "metadata": %s`, o.msgType, o.metadata)
}

func payloadsOperations(ops ...operation) []string {
	r := make([]string, len(ops))
	for i, op := range ops {
		r = append(r, fmt.Sprintf(`{"operation_identifier":{"index": %d}, %s}`, i, op.String()))
	}

	return r
}
