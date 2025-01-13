package systemtests

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type restClient struct {
	client            *resty.Client
	networkIdentifier string
}

func newRestClient(url, network, blockchain string) *restClient {
	client := resty.New().
		SetHostURL("http://" + url).
		SetHeaders(map[string]string{"Content-Type": "application/json", "Accept": "application/json"})
	return &restClient{
		client:            client,
		networkIdentifier: fmt.Sprintf(`"network_identifier":{"network":"%s","blockchain": "%s"}`, network, blockchain),
	}
}

func (c *restClient) block(block string) ([]byte, error) {
	res, err := c.client.R().
		SetBody(fmt.Sprintf(`{%s, "block_identifier":{"index":%s}}`, c.networkIdentifier, block)).
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
