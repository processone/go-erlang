package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestEncodeBinary(t *testing.T) {

}

func BenchmarkEncodeString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = bert.Encode("test")
	}
}

func BenchmarkBufferString(b *testing.B) {
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		_ = bert.EncodeTo(&buf, "test")
	}
}
