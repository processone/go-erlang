package bert_test

import (
	"bytes"
	"encoding/binary"
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
	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, 42); err != nil {
		t.Error(err)
	}
	if binary.BigEndian.Uint32(buf.Bytes()[1:5]) != 42 {
		t.Errorf("Unexpected int value")
	}
}

func BenchmarkBufferString(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		_ = bert.EncodeTo(&buf, "test")
	}
}
