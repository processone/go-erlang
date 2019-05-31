package bertrpc // import "gosrc.io/erlang/bertrpc"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

// Encode serializes a term as a ETF structure
func Encode(term interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := EncodeTo(term, &buf); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// Use Erlang External Term Format
// Reference: http://erlang.org/doc/apps/erts/erl_ext_dist.html
func EncodeTo(term interface{}, buf *bytes.Buffer) error {
	// Header for External Erlang Term Format
	buf.Write([]byte{TagETFVersion})

	// Encode the data
	if err := encodePayloadTo(term, buf); err != nil {
		return err
	}
	return nil
}

func encodePayloadTo(term interface{}, buf *bytes.Buffer) error {
	var err error
	switch t := term.(type) {

	case String:
		if t.ErlangType == StringTypeAtom {
			err = encodeAtom(buf, t.Value)
		} else {
			err = encodeString(buf, t.Value)
		}

	case string:
		err = encodeString(buf, t)

	case int:
		err = encodeInt(buf, int32(t))
	case int8:
		err = encodeInt(buf, int32(t))
	case int16:
		err = encodeInt(buf, int32(t))
	case int32:
		err = encodeInt(buf, t)
	case uint:
		err = encodeInt(buf, int32(t))
	case uint8:
		err = encodeInt(buf, int32(t))
	case uint16:
		err = encodeInt(buf, int32(t))
	case uint32:
		err = encodeInt(buf, int32(t))

	case Tuple:
		err = encodeTuple(buf, t)

	default:
		// Defines how to encode Go pointer types
		v := reflect.ValueOf(term)
		switch v.Kind() {
		case reflect.Slice:
			// TODO: handle reflect.Array
			var list []interface{}
			list, err = makeGenericSlice(term)
			if err != nil {
				err = fmt.Errorf("error converting slice: %v - %v:\n%v", v.Kind(), v.Type().Name(), err)
				break
			}
			err = encodeList(buf, list)
		default:
			err = fmt.Errorf("unhandled type: %v - %v", v.Kind(), v.Type().Name())
		}
	}
	return err
}

func encodeAtom(buf *bytes.Buffer, str string) error {
	// Encode atom header
	if len(str) <= 255 {
		// Encode small UTF8 atom
		buf.WriteByte(TagSmallAtomUTF8)
		buf.WriteByte(byte(len(str)))
	} else {
		// Encode standard UTF8 atom
		buf.WriteByte(TagAtomUTF8)
		if err := binary.Write(buf, binary.BigEndian, uint16(len(str))); err != nil {
			return err
		}
	}

	// Write atom
	buf.WriteString(str)
	return nil
}

func encodeString(buf *bytes.Buffer, str string) error {
	buf.WriteByte(TagBinary)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(str))); err != nil {
		return err
	}
	buf.WriteString(str)
	return nil
}

func encodeInt(buf *bytes.Buffer, i int32) error {
	if i >= 0 && i <= 255 {
		buf.WriteByte(TagSmallInteger)
		buf.WriteByte(byte(i))
	} else {
		buf.WriteByte(TagInteger)
		if err := binary.Write(buf, binary.BigEndian, i); err != nil {
			return err
		}
	}
	return nil
}

func encodeTuple(buf *bytes.Buffer, tuple Tuple) error {
	// Tuple header
	size := len(tuple.Elems)
	if size <= 255 {
		// Encode small tuple
		buf.WriteByte(TagSmallTuple)
		buf.WriteByte(byte(size))
	} else {
		// Encode large tuple
		buf.WriteByte(TagLargeTuple)
		if err := binary.Write(buf, binary.BigEndian, int32(size)); err != nil {
			return err
		}
	}

	// Tuple content
	for _, elem := range tuple.Elems {
		if err := encodePayloadTo(elem, buf); err != nil {
			return err
		}
	}
	return nil
}

func encodeList(buf *bytes.Buffer, list []interface{}) error {
	var err error
	// TODO: Special case for empty list: v.Len() ? Should not be needed

	// List header
	buf.WriteByte(TagList)
	if err := binary.Write(buf, binary.BigEndian, int32(len(list))); err != nil {
		return err
	}

	// List content
	for _, elem := range list {
		if err := encodePayloadTo(elem, buf); err != nil {
			return err
		}
	}
	// nil terminates the list:
	buf.Write([]byte{TagNil})
	return err
}

// ============================================================================
// Helpers

func makeGenericSlice(slice interface{}) ([]interface{}, error) {
	s := reflect.ValueOf(slice)
	switch s.Kind() {
	case reflect.Slice, reflect.Array:
		generic := make([]interface{}, s.Len())

		for i := 0; i < s.Len(); i++ {
			generic[i] = s.Index(i).Interface()
		}

		return generic, nil
	default:
		return []interface{}{},
			fmt.Errorf("cannot make a generic slice from something that is not a slice: %v", s.Kind())
	}
}
