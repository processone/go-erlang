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

func EncodeTo(buf *bytes.Buffer, term interface{}) error {
	var err error
	switch t := term.(type) {
	case Atom:
		err = encodeAtom(buf, t.Value)
	case string:
		err = encodeString(buf, t)
	case int:
		err = encodeInt(buf, int32(t))
	case int8:
		err = encodeInt(buf, int32(t))
	default:
		v := reflect.ValueOf(term)
		err = fmt.Errorf("unhandled type: %v", v.Kind())
	}
	return err
}

func encodeAtom(buf *bytes.Buffer, str string) error {
	// Encode atom header
	if len(str) <= 255 {
		// Encode small UTF8 atom
		buf.WriteByte(119) // TODO const
		buf.WriteByte(byte(len(str)))
	} else {
		// Encode standard UTF8 atom
		buf.WriteByte(118) // TODO const
		if err := binary.Write(buf, binary.BigEndian, uint16(len(str))); err != nil {
			return err
		}
	}

	// Write atom
	buf.WriteString(str)
	return nil
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

func encodeInt(buf *bytes.Buffer, i int32) error {
	buf.WriteByte(98) // TODO const
	if err := binary.Write(buf, binary.BigEndian, i); err != nil {
		return err
	}
	return nil
}
