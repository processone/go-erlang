package bert

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// A Bert call reply is either:
// {reply, Result}
// {error, {Type, Code, Class, Detail, Backtrace}}
func DecodeReply(r io.Reader, term interface{}) error {
	// 1. Read BERP length
	byte4 := make([]byte, 4)
	n, err := r.Read(byte4)
	if err != nil {
		return err
	}
	if n < 4 {
		return fmt.Errorf("truncated data")
	}
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
		res := FunctionResult{Result: term}
		if err := decodeData(r, &res); err != nil {
			return err
		}
		if res.Err != nil {
			return res.Err
		}

		return nil
	case "error":
		// TODO Decode Bert Error
		return errors.New("TODO Decode Bert error")
	default:
		return fmt.Errorf("incorrect reply tag: %s", tag)
	}
}

// Verify that we are reading a tuple and return the length of the tuple
func ReadTupleInfo(r io.Reader) (int, error) {
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

func ReadAtom(r io.Reader) (string, error) {
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
