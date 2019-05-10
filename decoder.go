package bert

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

func Decode2(r io.Reader, val interface{}) error {
	return nil
}

var ErrRange = errors.New("value out of range")

// TODO: Decode alternatives responses allow passing extra val as ...
func Decode(r io.Reader, term interface{}) error {
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	// Read Erlang Term Format "magic byte"
	if byte1[0] != byte(TagETFVersion) {
		// Bad Version tag (aka 'magic number')
		return errors.New("incorrect Erlang Term version tag")
	}

	// TODO: Test against valueof as Ptr
	val := reflect.ValueOf(term).Elem()
	switch val.Kind() {
	case reflect.Int8:
		return ErrRange
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := decodeInt(r)
		if err == nil {
			val.SetInt(i)
		}
		return err
	case reflect.String:
		s, err := decodeString(r)
		if err == nil {
			val.SetString(s)
		}
		return err
	default:
		return fmt.Errorf("unhandled decoding target")
	}
	return nil
}

// TODO: Pass bitsize here to trigger overflow operations errors
func decodeInt(r io.Reader) (int64, error) {
	// Read Tag
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return 0, err
	}

	// Compare expected type
	switch int(byte1[0]) {

	case TagSmallInteger:
		_, err = r.Read(byte1)
		if err != nil {
			return 0, err
		}
		return int64(byte1[0]), nil

	case TagInteger:
		byte4 := make([]byte, 4)
		n, err := r.Read(byte4)
		if err != nil {
			return 0, err
		}
		if n < 4 {
			return 0, fmt.Errorf("cannot decode integer, only %d bytes read", n)
		}
		var32 := int32(binary.BigEndian.Uint32(byte4))
		return int64(var32), nil
	}

	return 0, fmt.Errorf("incorrect type")
}

// We can decode several Erlang type in a string: Atom (Deprecated), AtomUTF8, Binary, CharList.
func decodeString(r io.Reader) (string, error) {
	// Read Tag
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return "", err
	}

	// Compare expected type
	switch int(byte1[0]) {

	case TagDeprecatedAtom, TagAtomUTF8, TagString:
		data, err := decodeString2(r)
		return string(data), err

	case TagSmallAtomUTF8:
		// Length:
		_, err = r.Read(byte1)
		if err != nil {
			return "", err
		}
		length := int(byte1[0])

		// Content:
		data := make([]byte, length)
		n, err := r.Read(data)
		if err != nil {
			return "", err
		}
		if n < length {
			return "", fmt.Errorf("truncated SmallAtomUTF8")
		}
		return string(data), nil

	}

	return "", fmt.Errorf("incorrect type")
}

func decodeString2(r io.Reader) ([]byte, error) {
	// Length:
	l := make([]byte, 2)
	_, err := r.Read(l)
	if err != nil {
		return []byte{}, err
	}
	length := int(binary.BigEndian.Uint16(l))

	// Content:
	data := make([]byte, length)
	n, err := r.Read(data)
	if err != nil {
		return []byte{}, err
	}
	if n < length {
		return []byte{}, fmt.Errorf("truncated data")
	}

	return data, nil
}
