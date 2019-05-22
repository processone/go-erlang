package bert

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
// If the term passed to the function is nil, we will not try to decode the function return.
func DecodeReply2(r io.Reader, term interface{}) error {
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
	length, err := ReadTupleInfo(r)
	if err != nil {
		return err
	}
	if length != 2 {
		return errors.New("unexpected bert reply tuple size")
	}

	// 4. Read the first Atom
	tag, err := ReadAtom(r)
	if err != nil {
		return err
	}

	// 5. Decode the reply or the error
	switch tag {
	case "reply":
		// Read the result of the function call
		if err := decodeData2(r, term); err != nil {
			return err
		}

		return nil
	case "error":
		// TODO Decode Bert Error
		return errors.New("TODO Decode Bert error")
	default:
		return fmt.Errorf("incorrect reply tag: %s", tag)
	}
}

// TODO ignore unexported fields
func decodeData2(r io.Reader, term interface{}) error {
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
		return decodeStruct2(r, val)

	default:
		return fmt.Errorf("unhandled decoding target: %s", val.Kind())
	}
}

func decodeStruct2(r io.Reader, val reflect.Value) error {
	// 1. Get the first field to determine, if we are decoding a tag value
	// It must be a string and be tagged as erlang:"tag"
	structType := val.Type()
	field1 := structType.Field(0)
	if tag, ok := field1.Tag.Lookup("erlang"); ok && tag == "tag" && field1.Type.Kind() == reflect.String {
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

	length := 0
	// TODO: split in two function to make it more readable
	switch int(byte1[0]) {
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
	}

	// An empty tuple cannot have a tag
	if length == 0 {
		return fmt.Errorf("tag expected in empty tuple")
	}

	// Extract first field as tag
	data, err := ReadAtom(r)
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
					err := decodeData2(r, currField.Addr().Interface())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

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
		return fmt.Errorf("cannot decode type %d to struct %s", int(byte1[0]), val.Type())
	}

	return decodeStructElts(r, length, val)
}
