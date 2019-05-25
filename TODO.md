# TODO

## Go module

+ Initial version for simple calls.
- Rename repository and module to 'erlang', 'erl', 'gerl' or 'goei' (for Go <-> Erlang Interface)
- Add support for slices / list
- Add support for BigInt
- Add support for Maps
- Make BERP header (4 byte length) optional. BERP header is not needed on HTTP, as framing will be done at HTTP level.
  However, I need to consider if I should always add it for consistency. It would also allow grouping several calls
  in a single HTTP request.
- Support zlib compression.
- Add Server example.
- Performance optimization.
- Support new error package (Go 2).
- Test and handle Erlang exceptions in function calls.
- Give the ability to test against protocol error vs Erlang returned errors.

## ejabberd_rpc module

- Support reading Erlang cookie to protect call behind bearer or basic auth
- Add ability to configure whitelist of modules that admin is allowed to call through RPC.
- Document configuration
- Support overloading cookie to have specific credential for that RPC endpoint.
- Add JWT token support
- Support BERT-RPC over MQTT
- TODO Improve help for clients: The client need to be able to retrieve a list of enabled modules to be able to display 
  proper help with available commands.
- High level Erlang API call: Allow using maps (or at least proplists). Make them easier to configure.
- Support Unix socket (example Erlang usage: https://stackoverflow.com/a/38286954/559289)

## Generic Erlang / Elixir TCP server

- Prepare Erlang module and ejabberd dependency: bert-server
- Use Ranch as a dependency or start from scratch to avoid dependencies?

## Examples

- Interop with Erlang and ejabberd
- Interop with Elixir
