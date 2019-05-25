package bert

import (
	"net/http"
)

type Client struct {
	Endpoint string
	// TODO make httpclient configurable
}

// TODO: Support getting token for authentication.
func New(endpoint string) Client {
	client := Client{Endpoint: endpoint}
	return client
}

type call struct {
	module   string
	function string
	args     []interface{}
}

func (Client) NewCall(module string, function string, args ...interface{}) call {
	return call{module: module, function: function, args: args}
}

func (c Client) Exec(call call, result interface{}) error {
	// Prepare BERT-RPC Packet
	buf, err := EncodeCall(call.module, call.function, call.args...)

	if err != nil {
		return err
	}

	// Post BERT-RPC call to HTTP endpoint
	resp, err := http.Post(c.Endpoint, "application/bert", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	/*
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println(body)
	*/

	return DecodeReply(resp.Body, result)
}
