package bert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// Atom is a wrapper structure to support Erlang atom data type.
type Atom struct {
	Value string
}

// Charlist is a wrapper structure to support Erlang charlist in encoding.
// Charlist is only used in encoding. On decoding, charlists are always decoded
// as strings.
type CharList struct {
	Value string
}

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

/*
func Encode(term interface{}) ([]byte, error) {
	v := reflect.ValueOf(term)
	switch v.Kind() {
	case reflect.String:
		encodeStringBinary(v.String())
	default:
		fmt.Printf("Unhandled type %v (%v)\n", v.Kind(), reflect.TypeOf(term).Elem().Name())
	}
	return []byte{}, nil
}

func encodeStringBinary(str string) []byte {
	headerLength := 5
	strLength := len(str)
	totalLength := headerLength + strLength
	hdr := make([]byte, totalLength)
	hdr[0] = 109
	binary.BigEndian.PutUint32(hdr[1:5], uint32(totalLength))
	return append(hdr, []byte(str)...)
}
*/

func EncodeTo(buf *bytes.Buffer, term interface{}) error {
	var err error
	switch t := term.(type) {
	case Atom:
		fmt.Println("This is an atom")
	case string:
		err = encodeString(buf, t)
	default:
		v := reflect.ValueOf(term)
		err = fmt.Errorf("unhandled type: %v", v.Kind())
	}
	return err
}

func encodeString(buf *bytes.Buffer, str string) error {
	// TODO make binary tag a constant
	buf.WriteByte(109)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(str))); err != nil {
		return err
	}
	buf.WriteString(str)
	return nil
}
