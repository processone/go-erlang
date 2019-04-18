package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestEncodeSmallAtom(t *testing.T) {
	atom := bert.Atom{Value: "atom"}
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
	var IntTests = []struct {
		n        int
		expected []byte
	}{
		// TODO: Include standalone header in that test 131.
		{-1, []byte{bert.TagInteger, 255, 255, 255, 255}},
		{1, []byte{bert.TagSmallInteger, 1}},
		{42, []byte{bert.TagSmallInteger, 42}},
		{255, []byte{bert.TagSmallInteger, 255}},
		{256, []byte{bert.TagInteger, 0, 0, 1, 0}},
		{1000, []byte{bert.TagInteger, 0, 0, 3, 232}},
	}

	for _, tt := range IntTests {
		var buf bytes.Buffer
		if err := bert.EncodeTo(&buf, tt.n); err != nil {
			t.Error(err)
		}
		if !bytes.Equal(buf.Bytes(), tt.expected) {
			t.Errorf("EncodeInt %d: expected %v, actual %v", tt.n, tt.expected, buf)
		}
	}
}

func BenchmarkBufferString(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		_ = bert.EncodeTo(&buf, "test")
	}
}
