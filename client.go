package bert

import (
	"io/ioutil"
	"net/http"
)

type Client struct {
	Endpoint string
	// TODO make httpclient configurable
}

func New(endpoint string) Client {
	client := Client{Endpoint: endpoint}
	return client
}

func (c Client) Call(module string, function string, params ...interface{}) (interface{}, error) {
	// Prepare BERT-RPC Packet
	buf, err := EncodeCall(module, function, params...)
	if err != nil {
		return nil, err
	}

	// Post BERT-RPC call to HTTP endpoint
	resp, err := http.Post(c.Endpoint, "application/bert", &buf)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	//DecodeResponse(body)

	return body, err
}
