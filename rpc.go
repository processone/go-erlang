package bert

import (
	"bytes"
	"encoding/binary"
)

// EncodeCall prepare a BERT-RPC Packet
// See: http://bert-rpc.org/
func EncodeCall(module string, function string, args ...interface{}) (bytes.Buffer, error) {
	var buf bytes.Buffer

	// -- {call, Module, Function, Arguments}
	call := T(A("call"), A(module), A(function), args)
	data, err := Marshal(call)
	if err != nil {
		return buf, err
	}

	// BERP Header = 4-bytes length
	// TODO: This should be optional for HTTP as it forces an extra allocation, instead of directly writing to the buffer
	//       We already have packet framing at the HTTP call level.
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(data))); err != nil {
		return buf, err
	}

	// Finally, write the data, after the length header
	buf.Write(data)
	return buf, err
}

// TODO: Use reader to read step by step from body
func DecodeResponse(resp []byte, patterns []interface{}) (interface{}, error) {
	return nil, nil
}
