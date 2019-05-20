package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestDecodeReply(t *testing.T) {
	// {reply, {error, exists}}
	input := []byte{0, 0, 0, 30, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 5, 101, 114, 114, 111,
		114, 100, 0, 6, 101, 120, 105, 115, 116, 115}

	var tuple struct {
		Data string
	}
	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &tuple)
	if err == nil {
		t.Errorf("bert decoding should have returned an error")
		return
	}

	if err.Error() != "exists" {
		t.Errorf("bert decoding should have returned error with reason 'exists'")
	}
}
