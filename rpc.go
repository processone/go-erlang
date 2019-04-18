package bert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// Use Erlang Term Format
// Reference: http://erlang.org/doc/apps/erts/erl_ext_dist.html
func EncodeCall(module string, function string, args ...interface{}) bytes.Buffer {
	var buffer bytes.Buffer
	// Header for External Erlang Terms
	buffer.Write([]byte{131})

	// We pass a tuple with 4 parameters
	buffer.Write([]byte{104, 4})

	// 1. atom(call)
	str := "call"
	buffer.Write([]byte{119, byte(len(str))})
	buffer.WriteString(str)

	// 2. atom(erlang_module)
	str = module
	buffer.Write([]byte{119, byte(len(str))})
	buffer.WriteString(str)

	// 3. atom(function)
	// Note: The function needs to be exported
	str = function
	buffer.Write([]byte{119, byte(len(str))})
	buffer.WriteString(str)

	// 4. Function Arguments [...]
	len4Bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(len4Bytes, uint32(len(args)))
	buffer.WriteByte(108)
	buffer.Write(len4Bytes)

	for _, arg := range args {
		fmt.Println(arg)
		v := reflect.ValueOf(arg)
		switch v.Kind() {
		case reflect.String:
			binary.BigEndian.PutUint32(len4Bytes, uint32(len(v.String())))
			buffer.WriteByte(109)
			buffer.Write(len4Bytes)
			buffer.WriteString(v.String())
		default:
			fmt.Printf("Unhandled type %v (%v)\n", v.Kind(), reflect.TypeOf(arg).Elem().Name())
		}
	}
	buffer.Write([]byte{106}) // nil terminates the list

	return buffer
}
