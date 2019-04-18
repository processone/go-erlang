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

const (
	TagSmallInteger  = 97
	TagInteger       = 98
	TagBinary        = 109
	TagAtomUTF8      = 118
	TagSmallAtomUTF8 = 119
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

// TODO Add support for small integer encoding
func encodeInt(buf *bytes.Buffer, i int32) error {
	buf.WriteByte(TagInteger)
	if err := binary.Write(buf, binary.BigEndian, i); err != nil {
		return err
	}
	return nil
}
