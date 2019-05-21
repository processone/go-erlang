# TODO

- Initial version for simple calls.
- Prepare Erlang module and ejabberd dependency: bert-server
- Add support for slices / list
- Add support for BigInt
- Add support for Maps
- Add BERP header (4 byte length). BERP header is not needed on HTTP, as framing will be done at HTTP level.
  However, I need to consider if I should always add it for consistency. It would also allow grouping several calls
  in a single HTTP request.
- Support zlib compression.
- Add Server example.
- Performance optimization.
- Support new error package (Go 2).
- Test and handle Erlang exceptions in function calls.
- Give the ability to test against protocol error vs Erlang returned errors.

## Examples

- Interop with Erlang and ejabberd
- Interop with Elixir
