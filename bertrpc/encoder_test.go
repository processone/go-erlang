package bertrpc_test // import "gosrc.io/erlang/bertrpc_test"

import (
	"bytes"
	"testing"

	"gosrc.io/erlang/bertrpc"
)

func TestEncodeSmallAtom(t *testing.T) {
	atom := bertrpc.A("atom")
	data, err := bertrpc.Marshal(atom)
	if err != nil {
		t.Error(err)
	}
	// Use new utf8 atom (119), instead of old deprecated atom (100)
	expected := []byte{131, 119, 4, 97, 116, 111, 109}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeSmallAtom: expected %v, actual %v", expected, data)
	}
}

// We encode strings to binary, but we can force them to charlist (see TestEncodeCharList)
func TestEncodeString(t *testing.T) {
	data, err := bertrpc.Marshal("string")
	if err != nil {
		t.Error(err)
	}
	expected := []byte{131, 109, 0, 0, 0, 6, 115, 116, 114, 105, 110, 103}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeString: expected %v, actual %v", expected, data)
	}
}

func TestEncodeInt(t *testing.T) {
	var tests = []struct {
		n        int
		expected []byte
	}{
		{-2147483648, []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 128, 0, 0, 0}},
		{-1, []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 255, 255, 255, 255}},
		{1, []byte{bertrpc.TagETFVersion, bertrpc.TagSmallInteger, 1}},
		{42, []byte{bertrpc.TagETFVersion, bertrpc.TagSmallInteger, 42}},
		{255, []byte{bertrpc.TagETFVersion, bertrpc.TagSmallInteger, 255}},
		{256, []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 0, 0, 1, 0}},
		{1000, []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 0, 0, 3, 232}},
		{2147483647, []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 127, 255, 255, 255}},
	}

	for _, tt := range tests {
		data, err := bertrpc.Marshal(tt.n)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(data, tt.expected) {
			t.Errorf("EncodeInt %d: expected %v, actual %v", tt.n, tt.expected, data)
		}
	}
}

func TestEncodeMiscInt(t *testing.T) {
	var tests = []struct {
		n        interface{}
		expected []byte
	}{
		// TODO: Include standalone header in that test 131.
		{int16(-256), []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 255, 255, 255, 0}},
		{int8(-1), []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 255, 255, 255, 255}},
		{uint8(1), []byte{bertrpc.TagETFVersion, bertrpc.TagSmallInteger, 1}},
		{uint16(256), []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 0, 0, 1, 0}},
		{uint32(2147483647), []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 127, 255, 255, 255}},
		{int32(2147483647), []byte{bertrpc.TagETFVersion, bertrpc.TagInteger, 127, 255, 255, 255}},
	}

	for _, tt := range tests {
		data, err := bertrpc.Marshal(tt.n)
		if err != nil {
			t.Error(err)
		}
		if !bytes.Equal(data, tt.expected) {
			t.Errorf("EncodeMiscInt %d: expected %v, actual %v", tt.n, tt.expected, data)
		}
	}
}

func TestEncodeTuple(t *testing.T) {
	tuple := bertrpc.T(bertrpc.A("atom"), "string", 42)

	data, err := bertrpc.Marshal(tuple)
	if err != nil {
		t.Error(err)
	}

	// In Erlang, generated tuple was using deprecated atom header 100,0,4 instead of 119, 4 (right after 104, 3)
	// However, the new version decodes just fine.
	// TODO: Deserialization should support deprecated header decoding
	expected := []byte{131, 104, 3, 119, 4, 97, 116, 111, 109, 109, 0, 0, 0, 6, 115, 116, 114, 105, 110, 103, 97, 42}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeTuple: expected %v, actual %v", expected, data)
	}
}

func TestEncodeLargeTuple(t *testing.T) {
	var els []interface{}

	for el := 0; el < 256; el++ {
		els = append(els, el)
	}
	tuple := bertrpc.Tuple{els}

	data, err := bertrpc.Marshal(tuple)
	if err != nil {
		t.Error(err)
	}

	// Inspect header
	expected := []byte{131, 105, 0, 0, 1, 0}
	header := data[0:6]
	if !bytes.Equal(header, expected) {
		t.Errorf("EncodeLargeTuple: expected %v, actual %v", expected, header)
	}
}

func TestEncodeList(t *testing.T) {
	list := bertrpc.L(bertrpc.A("atom"), "string", 42)

	data, err := bertrpc.Marshal(list)
	if err != nil {
		t.Error(err)
	}

	// Use new utf8 atom (119), instead of old deprecated atom (100)
	expected := []byte{131, 108, 0, 0, 0, 3, 119, 4, 97, 116, 111, 109, 109, 0, 0, 0, 6, 115,
		116, 114, 105, 110, 103, 97, 42, 106}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeList: expected %v, actual %v", expected, data)
	}
}

func TestEncodeIntSlice(t *testing.T) {
	list := []int{1, 2, 3}

	data, err := bertrpc.Marshal(list)
	if err != nil {
		t.Error(err)
	}

	// TODO: (most like for decoding). Erlang optimize this as string (list char): 131, 107, 0, 3, 1, 2, 3
	expected := []byte{131, 108, 0, 0, 0, 3, 97, 1, 97, 2, 97, 3, 106}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeIntSlice: expected %v, actual %v", expected, data)
	}
}

// Recursive structure: puts a list into a tuple
func TestEncodeTupleList(t *testing.T) {
	tuple := bertrpc.T(bertrpc.L(bertrpc.A("atom"), "string", 42))
	data, err := bertrpc.Marshal(tuple)
	if err != nil {
		t.Error(err)
	}
	// Use new utf8 atom (119), instead of old deprecated atom (100)
	expected := []byte{131, 104, 1, 108, 0, 0, 0, 3, 119, 4, 97, 116, 111, 109, 109, 0, 0, 0,
		6, 115, 116, 114, 105, 110, 103, 97, 42, 106}
	if !bytes.Equal(data, expected) {
		t.Errorf("EncodeTuple: expected %v, actual %v", expected, data)
	}
}

func BenchmarkBufferString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = bertrpc.Marshal("test")
	}
}
