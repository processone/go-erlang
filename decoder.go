package bert

import (
	"bytes"
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

	return decodeData(r, term)
}

// TODO ignore unexported fields
func decodeData(r io.Reader, term interface{}) error {
	// TODO: Test against valueof as not a Ptr
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
	case reflect.Struct:
		// Check if we are trying to get a function call return (special FunctionCall struct)
		if val.Type().Name() == "FunctionResult" {
			return decodeFunctionResult(r, val)
		}
		// Otherwise, decode directly to user-defined struct
		return decodeStruct(r, val)

	default:
		return fmt.Errorf("unhandled decoding target")
	}
}

func decodeStruct(r io.Reader, val reflect.Value) error {
	// For now we assume the structure and Erlang type is composed only of basic types
	// 1. Get the Erlang type of the structure:
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	// 2. Check that this is a tuple of same length
	length := 0
	switch int(byte1[0]) {
	case TagSmallTuple:
		_, err := r.Read(byte1)
		if err != nil {
			return err
		}
		length = int(byte1[0])
	case TagLargeTuple:
		byte4 := make([]byte, 4)
		n, err := r.Read(byte4)
		if err != nil {
			return err
		}
		if n < 4 {
			return fmt.Errorf("truncated data")
		}
		length = int(binary.BigEndian.Uint32(byte4))

	default:
		return fmt.Errorf("cannot decode type %d to struct", int(byte1[0]))
	}
	// If the tuple does not contain the expected number of fields in our struct
	if length != val.NumField() {
		return fmt.Errorf("cannot decode tuple of length %d to struct", length)
	}

	// For each field, try to decode it recursively
	for i := 0; i < length; i++ {
		valueField := val.Field(i)
		if valueField.Kind() == reflect.Ptr {
			valueField = valueField.Elem()
		}
		if valueField.CanAddr() {
			err = decodeData(r, valueField.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeFunctionResult(r io.Reader, val reflect.Value) error {
	// Read the type of data
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	tupleLength := 0
	switch int(byte1[0]) {

	// ========================================================================
	// Function return is an atom
	case TagDeprecatedAtom, TagAtomUTF8:
		data, err := decodeString2(r)
		if err != nil {
			return err
		}

		// return is ok => Success = true
		if bytes.Compare(data, []byte{111, 107}) == 0 {
			valueField := val.FieldByName("Success")
			valueField.SetBool(true)
			return nil
		}

		// return is error => Set error to generic value
		if bytes.Compare(data, []byte{101, 114, 114, 111, 114}) == 0 {
			valueField := val.FieldByName("Err")
			valueField.Set(reflect.ValueOf(ErrReturn))
			return nil
		}

		// return is not atom or or error. If the expect result is a string, we consider the return the result.
		// TODO: Make more generic
		valueField := val.FieldByName("Result")
		embeddedVal := valueField.Interface()
		nv := reflect.ValueOf(embeddedVal).Elem()
		if nv.Kind() == reflect.String {
			nv.SetString(string(data))
			return nil
		}
		// Otherwise we fail.

		return fmt.Errorf("unexpected result type in FunctionResult")

	// ========================================================================
	// Function return is a tuple
	case TagSmallTuple:
		_, err := r.Read(byte1)
		if err != nil {
			return err
		}
		tupleLength = int(byte1[0])
	case TagLargeTuple:
		byte4 := make([]byte, 4)
		n, err := r.Read(byte4)
		if err != nil {
			return err
		}
		if n < 4 {
			return fmt.Errorf("truncated data")
		}
		tupleLength = int(binary.BigEndian.Uint32(byte4))

	default:
		return fmt.Errorf("cannot decode type %d to struct", int(byte1[0]))
	}

	// Decode tuple. If tuple if of length 2, check if it is either {ok, Result} or {error, Reason}.
	// Otherwise, consider it to be the result.

	fmt.Println(tupleLength)
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

	case TagBinary:
		data, err := decodeString4(r)
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

	return "", fmt.Errorf("incorrect type")
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
	if err != nil {
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
	if err != nil {
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
