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

type Tuple struct {
	Elems []interface{}
}

// Charlist is a wrapper structure to support Erlang charlist in encoding.
// Charlist is only used in encoding. On decoding, charlists are always decoded
// as strings.
type CharList struct {
	Value string
}

// Short factory functions to help write short structure generation code.
func A(str string) Atom {
	return Atom{str}
}

func T(el ...interface{}) Tuple {
	return Tuple{el}
}

// Supported types
const (
	TagSmallInteger  = 97
	TagInteger       = 98
	TagSmallTuple    = 104
	TagLargeTuple    = 105
	TagBinary        = 109
	TagAtomUTF8      = 118
	TagSmallAtomUTF8 = 119
	TagETFVersion    = 131
)

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
		v := reflect.ValueOf(term)
		err = fmt.Errorf("unhandled type: %v - %v", v.Kind(), v.Type().Name())
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
	buf.WriteByte(TagSmallTuple)
	buf.WriteByte(byte(len(tuple.Elems)))
	for _, elem := range tuple.Elems {
		if err := EncodeTo(buf, elem); err != nil {
			return err
		}
	}
	return nil
}
