# TODO

- Initial version for simple calls.
- Add BERP header (4 byte length). BERP header is not needed on HTTP, as framing will be done at HTTP level.
  However, I need to consider if I should always add it for consistency. It would also allow grouping several calls
  in a single HTTP request.