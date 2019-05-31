package bertrpc // import "gosrc.io/erlang/bertrpc"

import (
	"net/http"
)

// Client create an HTTP client to that holds configuration parameters to make Bert-RPC calls.
type Client struct {
	// This is the endpoint used to access bert-rpc server
	// For now, we only support HTTP endpoints.
	Endpoint string
	// This is the security token used to pass call (HTTP bearer token auth)
	Token string

	// TODO: make httpclient configurable
}

// TODO: Support getting token for authentication.
func New(endpoint string) Client {
	client := Client{Endpoint: endpoint}
	return client
}

// call is the internal structure to hold bert-rpc call parameters
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

	// Use HTTP POST to trigger BERT-RPC call over HTTP
	resp, err := http.Post(c.Endpoint, "application/bert", &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return DecodeReply(resp.Body, result)
}
