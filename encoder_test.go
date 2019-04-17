package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestEncodeAtom(t *testing.T) {
	atom := bert.Atom{Value: "atom"}
	var buf bytes.Buffer
	if err := bert.EncodeTo(&buf, atom); err != nil {
		t.Error(err)
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
}

/*
func BenchmarkEncodeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = bert.Encode("test")
	}
}
*/

func BenchmarkBufferString(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		_ = bert.EncodeTo(&buf, "test")
	}
}
