// Example client showing how to use BERT-RPC to create a user in ejabberd.
package main

import (
	"log"

	"gosrc.io/erlang/bertrpc"
)

func main() {
	svc := bertrpc.New("http://localhost:5281/rpc/")
	c := svc.NewCall("ejabberd_auth", "try_register", "john", "localhost", "password")
	var result struct { // ok | {error, atom()}
		Tag    string `erlang:"tag"`
		Reason string `erlang:"tag:error"`
	}
	err := svc.Exec(c, &result)
	if err != nil {
		// Protocol or decoding errors
		log.Fatal("operation failed: ", err)
	}

	switch result.Tag {
	case "ok":
		log.Println("Successfully created user")
	case "error":
		log.Fatal("Could not create user: ", result.Reason)
	}
}

/*
This module assumes that ejabberd has been configured with ejabberd_rpc support. This module is available in ejabberd
master repository.

Example config:

```
# Listener. ejabberd bertrpc module will be available on localhost on port 5281, under /rpc/ http endpoint.
listen:
# ...
  -
    port: 5281
    # For IPv6, use:
    # ip: "::FFFF:127.0.0.1"
    ip:  "127.0.0.1"
    module: ejabberd_http
    request_handlers:
      "rpc": ejabberd_rpc
```
*/
