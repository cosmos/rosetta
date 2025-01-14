package systemtests

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
		SetHostURL("http://" + rosetta.Addr).
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
	accountIdentifier := fmt.Sprintf(`"account_identifier":{"address": "%s"}`, address)

	bodyOpts := make([]string, len(opts))
	for i, opt := range opts {
		bodyOpts[i] = opt()
	}

	args := []string{c.networkIdentifier, accountIdentifier}
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
