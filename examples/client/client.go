// Example client showing how to use BERT-RPC to create a user in ejabberd.
package main

import (
	"fmt"

	"github.com/processone/bert"
)

func main() {
	svc := bert.New("http://localhost:5281/rpc/")
	body, _ := svc.Call("ejabberd_auth", "try_register",
		"john", "localhost", "password")
	fmt.Println(body)
}

/*
This module assumes that ejabberd has been configured with mod_bertrpc support. This module is available starting
from ejabberd 19.05.

Example config:

# Listener. bertrpc module will be available on localhost ipv6 on port 5281, under /rpc/ http endpoint.
listen:
# ...
  -
    port: 5281
    ip: "::FFFF:127.0.0.1"
    module: ejabberd_http
    request_handlers:
      "rpc": mod_bertrpc

*/
