package bertrpc // import "gosrc.io/erlang/bertrpc"

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// A Bert call reply is either:
// {reply, Result}
// {error, {Type, Code, Class, Detail, Backtrace}}
// If we pass an empty struct it means we do not care about the reply and we will not try to decode
// Erlang return.
func DecodeReply(r io.Reader, term interface{}) error {
	// Guard against nil decoding target  as it does not guide the decoding
	if term == nil {
		return fmt.Errorf("target type for decoding cannot be nil")
	}

	// 1. Read BERP length
	byte4 := make([]byte, 4)
	n, err := r.Read(byte4)
	if err != nil {
		return err
	}
	if n < 4 {
		return fmt.Errorf("truncated data")
	}
	// TODO: Keep track of the length of the data read, to be able to skip to the end on failure.
	_ = int(binary.BigEndian.Uint32(byte4))

	// 2. Read Erlang Term Format "magic byte"
	byte1 := make([]byte, 1)
	_, err = r.Read(byte1)
	if err != nil {
		return err
	}
	if byte1[0] != byte(TagETFVersion) {
		// Bad Version tag (aka 'magic number')
		return fmt.Errorf("incorrect Erlang Term version tag: %d", byte1[0])
	}

	// 3. Read the reply tuple header
	length, err := readTupleInfo(r)
	if err != nil {
		return err
	}
	if length != 2 {
		return errors.New("unexpected bert reply tuple size")
	}

	// 4. Read the first Atom
	tag, err := readAtom(r)
	if err != nil {
		return err
	}

	// 5. Decode the reply or the error
	switch tag {
	case "reply":
		// Read the result of the function call
		if err := decodeData(r, term); err != nil {
			return err
		}

		return nil
	case "error":
		// TODO Decode Bert Error and add test on errors
		return errors.New("TODO Decode Bert error")
	default:
		return fmt.Errorf("incorrect reply tag: %s", tag)
	}
}

// ============================================================================
// Decode Erlang Term format into a Go structure

// TODO ignore unexported fields
func decodeStruct(r io.Reader, val reflect.Value) error {
	// If the struct is empty, we assume caller is not interested in the result
	// and we do not try to decode anything.
	if val.NumField() == 0 {
		return nil
	}

	// Get the first field of the interface we are decoding to, to determine
	// if we are decoding a target value.
	// It must be a string and be tagged as erlang:"tag"
	structType := val.Type()
	field1 := structType.Field(0)
	tag, ok := field1.Tag.Lookup("erlang")
	if ok && tag == "tag" && field1.Type.Kind() == reflect.String {
		return decodeTaggedValue(r, val)
	}
	return decodeUntaggedStruct(r, val)
}

func decodeTaggedValue(r io.Reader, val reflect.Value) error {
	// We need to read Erlang data type. If we have an atom, it will be the tag.
	// If we have a tuple, We expect first element to be the tag.
	// If we have something else, we try to decode it in an untagged field.
	// Read the type of data
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

	switch int(byte1[0]) {
	// We are directly decoding the tag, return it inside the struct:
	case TagDeprecatedAtom, TagAtomUTF8, TagSmallAtomUTF8:
		return readTagAtom(r, int(byte1[0]), val)
	case TagSmallTuple, TagLargeTuple:
		return readTagTuple(r, int(byte1[0]), val)
	}
	/*
		// If the data is not an atom nor a tuple, we decode in the next data structure that is not associated to a tag
		// Searching for a freeform raw field
		structType := val.Type()
		for i := 1; i < structType.NumField(); i++ {
			field := structType.Field(i)
			t, _ := field.Tag.Lookup("erlang")
			// Field is a tagged value, we skip it
			if strings.HasPrefix(t, "tag:") {
				continue
			}

			// We found a candidate field for decoding
			currField := val.Field(i)
			if currField.Kind() == reflect.Ptr {
				currField = currField.Elem()
			}
			if currField.CanAddr() {
				return readOtherData(r, int(byte1[0]), currField.Addr().Interface())
			}
		}
	*/
	// We did not find any field to decode the tag to
	return fmt.Errorf("decodeTaggedValue could not read atom or taggedTuple")
}

func readTagAtom(r io.Reader, erlangType int, val reflect.Value) error {
	switch erlangType {
	// We are directly decoding the tag, return it inside the struct:
	case TagDeprecatedAtom, TagAtomUTF8:
		data, err := decodeString2(r)
		if err != nil {
			return err
		}
		field1 := val.Field(0)
		field1.SetString(string(data))
		return nil
	case TagSmallAtomUTF8:
		data, err := decodeString1(r)
		if err != nil {
			return err
		}
		field1 := val.Field(0)
		field1.SetString(string(data))
		return nil
	default:
		return fmt.Errorf("readTagAtom unexpected mismatch: %d", erlangType)
	}
}

func readTagTuple(r io.Reader, erlangType int, val reflect.Value) error {
	// Get tuple length
	byte1 := make([]byte, 1)
	length := 0
	switch erlangType {
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
		return fmt.Errorf("readTagTuple unexpected mismatch: %d", erlangType)
	}

	// An empty tuple cannot have a tag
	if length == 0 {
		return fmt.Errorf("tag cannot be found in an empty tuple")
	}

	// Extract first field as tag
	data, err := readAtom(r)
	tag := string(data)
	if err != nil {
		return fmt.Errorf("cannot read atom as first tuple element")
	}
	field1 := val.Field(0)
	field1.SetString(tag)

	// Match all others fields against the tag name constraint to decode the fields one by one
	structType := val.Type()
	for i := 1; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if t, ok := field.Tag.Lookup("erlang"); ok {
			if t == "tag:"+tag {
				currField := val.Field(i)
				if currField.Kind() == reflect.Ptr {
					currField = currField.Elem()
				}
				if currField.CanAddr() {
					err := decodeData(r, currField.Addr().Interface())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

/*
func readOtherData(r io.Reader, erlangType int, val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {

	case reflect.Int8:
		return ErrRange
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := decodeInt(r) // TODO Point to partial decodeInt, passing the Erlang type that was already read
		if err == nil {
			val.SetInt(i)
		}
		return err
	case reflect.String:
		s, err := decodeString(r) // TODO Point to partial decodeString, passing the Erlang type that was already read
		if err == nil {
			val.SetString(s)
		}
		return err

	default:
		return fmt.Errorf("readOtherData unexpected mismatch: %s", val.Kind())
	}
}
*/

// ============================================================================

func decodeUntaggedStruct(r io.Reader, val reflect.Value) error {
	// 1. Get the Erlang type of the tuple
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("cannot decode type %s to struct %s", erlangType(int(byte1[0])), val.Type())
	}

	return decodeStructElts(r, length, val)
}

func decodeStructElts(r io.Reader, length int, val reflect.Value) error {
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
			err := decodeData(r, valueField.Addr().Interface())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ============================================================================
// Helpers

// Verify that we are reading a tuple and return the length of the tuple
func readTupleInfo(r io.Reader) (int, error) {
	// 1. Read the type of data
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return 0, err
	}

	// 2. Return
	tupleLength := 0
	switch int(byte1[0]) {
	case TagSmallTuple:
		_, err := r.Read(byte1)
		if err != nil {
			return 0, err
		}
		tupleLength = int(byte1[0])
	case TagLargeTuple:
		byte4 := make([]byte, 4)
		n, err := r.Read(byte4)
		if err != nil {
			return 0, err
		}
		if n < 4 {
			return 0, fmt.Errorf("truncated data")
		}
		tupleLength = int(binary.BigEndian.Uint32(byte4))

	default:
		return 0, fmt.Errorf("cannot decode type %d to struct", int(byte1[0]))
	}

	return tupleLength, nil
}

func readAtom(r io.Reader) (string, error) {
	// Read the type of data
	byte1 := make([]byte, 1)
	_, err := r.Read(byte1)
	if err != nil {
		return "", err
	}

	switch int(byte1[0]) {
	case TagDeprecatedAtom, TagAtomUTF8:
		data, err := decodeString2(r)
		if err != nil {
			return "", err
		}
		return string(data), nil
	case TagSmallAtomUTF8:
		data, err := decodeString1(r)
		if err != nil {
			return "", err
		}
		return string(data), nil

	default:
		return "", fmt.Errorf("cannot decode type %d as atom", int(byte1[0]))
	}
}
