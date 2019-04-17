package bert

import (
	"fmt"
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

func (c Client) Call(module string, function string, params ...interface{}) interface{} {
	buf := EncodeCall(module, function, params...)
	resp, err := http.Post(c.Endpoint, "application/bert", &buf)
	if err != nil {
		fmt.Println(err)
		return "error"
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(body)
	return "ok"
}
