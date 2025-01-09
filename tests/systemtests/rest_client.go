package systemtests

import (
	"fmt"
	"github.com/go-resty/resty/v2"
)

type RestClient struct {
	client            *resty.Client
	networkIdentifier string
}

func newRestClient(url, network, blockchain string) *RestClient {
	client := resty.New().
		SetHostURL("http://" + url).
		SetHeaders(map[string]string{"Content-Type": "application/json", "Accept": "application/json"})
	return &RestClient{
		client:            client,
		networkIdentifier: fmt.Sprintf(`"network_identifier":{"network":"%s","blockchain": "%s"}`, network, blockchain),
	}
}

func (c *RestClient) block(block string) ([]byte, error) {
	res, err := c.client.R().
		SetBody(fmt.Sprintf(`{%s, "block_identifier":{"index":%s}}`, c.networkIdentifier, block)).
		Post("/block")
	if err != nil {
		return nil, err
	}
	return res.Body(), nil
}
