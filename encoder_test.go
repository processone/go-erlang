package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestEncodeSmallAtom(t *testing.T) {
	atom := bert.A("atom")
	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, atom); err != nil {
		t.Error(err)
	}
	result := append([]byte{119, byte(len(atom.Value))}, []byte(atom.Value)...)
	if !bytes.Equal(buf.Bytes(), result) {
		t.Errorf("Unexpected encoding")
	}
}

// We encode string to binary
func TestEncodeString(t *testing.T) {
	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, "string"); err != nil {
		t.Error(err)
	}
}

func TestEncodeInt(t *testing.T) {
	var tests = []struct {
		n        int
		expected []byte
	}{
		// TODO: Include standalone header in that test 131.
		{-2147483648, []byte{bert.TagInteger, 128, 0, 0, 0}},
		{-1, []byte{bert.TagInteger, 255, 255, 255, 255}},
		{1, []byte{bert.TagSmallInteger, 1}},
		{42, []byte{bert.TagSmallInteger, 42}},
		{255, []byte{bert.TagSmallInteger, 255}},
		{256, []byte{bert.TagInteger, 0, 0, 1, 0}},
		{1000, []byte{bert.TagInteger, 0, 0, 3, 232}},
		{2147483647, []byte{bert.TagInteger, 127, 255, 255, 255}},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		if err := bert.EncodeTo(&buf, tt.n); err != nil {
			t.Error(err)
		}
		if !bytes.Equal(buf.Bytes(), tt.expected) {
			t.Errorf("EncodeInt %d: expected %v, actual %v", tt.n, tt.expected, buf)
		}
	}
}

func TestEncodeMiscInt(t *testing.T) {
	var tests = []struct {
		n        interface{}
		expected []byte
	}{
		// TODO: Include standalone header in that test 131.
		{int16(-256), []byte{bert.TagInteger, 255, 255, 255, 0}},
		{int8(-1), []byte{bert.TagInteger, 255, 255, 255, 255}},
		{uint8(1), []byte{bert.TagSmallInteger, 1}},
		{uint16(256), []byte{bert.TagInteger, 0, 0, 1, 0}},
		{uint32(2147483647), []byte{bert.TagInteger, 127, 255, 255, 255}},
		{int32(2147483647), []byte{bert.TagInteger, 127, 255, 255, 255}},
	}

	for _, tt := range tests {
		var buf bytes.Buffer
		if err := bert.EncodeTo(&buf, tt.n); err != nil {
			t.Error(err)
		}
		if !bytes.Equal(buf.Bytes(), tt.expected) {
			t.Errorf("EncodeMiscInt %d: expected %v, actual %v", tt.n, tt.expected, buf)
		}
	}
}

func TestEncodeTuple(t *testing.T) {
	tuple := bert.T(bert.A("atom"), "string", 42)

	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, tuple); err != nil {
		t.Error(err)
	}

	// In Erlang, generated tuple was using deprecated atom header 100,0,4 instead of 119, 4 (right after 104, 3)
	// However, the new version decodes just fine.
	// TODO: Deserialization should support deprecated header decoding
	expected := []byte{104, 3, 119, 4, 97, 116, 111, 109, 109, 0, 0, 0, 6, 115, 116, 114, 105, 110, 103, 97, 42}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("EncodeTuple: expected %v, actual %v", expected, buf)
	}
}

func TestEncodeLargeTuple(t *testing.T) {
	var els []interface{}

	for el := 0; el < 256; el++ {
		els = append(els, el)
	}
	tuple := bert.Tuple{els}

	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, tuple); err != nil {
		t.Error(err)
	}

	// Inspect header
	expected := []byte{105, 0, 0, 1, 0}
	header := buf.Bytes()[0:5]
	if !bytes.Equal(header, expected) {
		t.Errorf("EncodeTuple: expected %v, actual %v", expected, header)
	}
}

func TestEncodeList(t *testing.T) {
	list := bert.L(bert.A("atom"), "string", 42)

	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, list); err != nil {
		t.Error(err)
	}
	// Use new utf8 atom (119), instead of old deprecated atom (100)
	expected := []byte{108, 0, 0, 0, 3, 119, 4, 97, 116, 111, 109, 109, 0, 0, 0, 6, 115,
		116, 114, 105, 110, 103, 97, 42, 106}
	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("EncodeTuple: expected %v, actual %v", expected, buf.Bytes())
	}
}

func TestEncodeIntSlice(t *testing.T) {
	list := []int{1, 2, 3}

	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, list); err != nil {
		t.Error(err)
	}
}

func BenchmarkBufferString(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		_ = bert.EncodeTo(&buf, "test")
	}
}
