package bert

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

/*
TODO: Change the approach ? Fully decode the structure recursively and try to map it to the target type.
   It might help handle more complex structure, with several level of embedded structure.
*/

var ErrRange = errors.New("value out of range")
var ErrReturn = errors.New("function returns 'error'")

/*
Special type to decode function call returns.

In general, a function call in Erlang returns one of the following structure:
- ok
- error
- {ok, Result}
- {error, Reason}
- Result

It could also throw an error to end in error.

This special type is used to be able to map it with Go convention, either
getting a valid result or an error.
*/
type FunctionResult struct {
	Success bool
	Err     error
	Result  interface{}
}

func Decode(r io.Reader, term interface{}) error {
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	// Read Erlang Term Format "magic byte"
	if byte1[0] != byte(TagETFVersion) {
		// Bad Version tag (aka 'magic number')
		return fmt.Errorf("incorrect Erlang Term version tag: %d", byte1[0])
	}

	return decodeData(r, term)
}

// TODO ignore unexported fields
func decodeData(r io.Reader, term interface{}) error {
	// Resolve pointers
	val := reflect.ValueOf(term)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

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
	case reflect.Struct:
		return decodeStruct(r, val)

	default:
		return fmt.Errorf("unhandled decoding target: %s", val.Kind())
	}
}

// ============================================================================
// Decode basic types

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

// We can decode several Erlang types in a string: Atom (Deprecated), AtomUTF8, Binary, CharList.
func decodeString(r io.Reader) (string, error) {
	// Read Tag
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return "", err
	}

	// Compare expected type
	dataType := int(byte1[0])
	switch dataType {

	case TagSmallAtomUTF8:
		data, err := decodeString1(r)
		return string(data), err

	case TagDeprecatedAtom, TagAtomUTF8, TagString:
		data, err := decodeString2(r)
		return string(data), err

	case TagBinary:
		data, err := decodeString4(r)
		return string(data), err

	case TagList:

		// Count:
		byte4 := make([]byte, 4)
		n, err := r.Read(byte4)
		if err != nil {
			return "", err
		}
		if n < 4 {
			return "", fmt.Errorf("truncated List data")
		}
		count := int(binary.BigEndian.Uint32(byte4))

		s := []rune("")
		// Last element in list should be termination marker, so we loop (count - 1) times
		for i := 1; i <= count; i++ {
			// Assumption: We are decoding a into a string, so we expect all elements to be integers;
			// We can fail otherwise.
			char, err := decodeInt(r)
			if err != nil {
				return "", err
			}
			// Erlang does not encode utf8 charlist into a series of bytes, but use large integers.
			// We need to process the integer list as runes.
			s = append(s, rune(char))
		}
		// TODO: Check that we have the list termination mark
		if err := decodeNil(r); err != nil {
			return string(s), err
		}

		return string(s), nil
	}

	return "", fmt.Errorf("incorrect type: %d", dataType)
}

func decodeString1(r io.Reader) ([]byte, error) {
	// Length:
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return []byte{}, err
	}
	length := int(byte1[0])

	// Content:
	data := make([]byte, length)
	n, err := r.Read(data)
	if err != nil && err != io.EOF {
		return []byte{}, err
	}
	if n < length {
		return []byte{}, fmt.Errorf("truncated data")
	}
	return data, nil

}

// Decode a string with length on 16 bits.
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
	if err != nil && err != io.EOF {
		return []byte{}, err
	}
	if n < length {
		return []byte{}, fmt.Errorf("truncated data")
	}

	return data, nil
}

// Decode a string with length on 32 bits.
func decodeString4(r io.Reader) ([]byte, error) {
	// Length:
	l := make([]byte, 4)
	_, err := r.Read(l)
	if err != nil {
		return []byte{}, err
	}
	length := int(binary.BigEndian.Uint32(l))

	// Content:
	data := make([]byte, length)
	n, err := r.Read(data)
	if err != nil && err != io.EOF {
		return []byte{}, err
	}
	if n < length {
		return []byte{}, fmt.Errorf("truncated data")
	}

	return data, nil
}

// Read a nil value and return error in case of unexpected value.
// Nil is expected as a marker for end of lists.
func decodeNil(r io.Reader) error {
	// Read Tag
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	if byte1[0] != byte(TagNil) {
		return fmt.Errorf("could not find nil: %d", byte1[0])
	}

	return nil
}
