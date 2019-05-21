package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

func TestDecodeErrorReply(t *testing.T) {
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
		t.Errorf("bert decoding should have returned error with reason 'exists': %s", err.Error())
	}
}

func TestDecodeOkReply(t *testing.T) {
	// {reply, {ok, 110}}
	input := []byte{0, 0, 0, 20, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 2, 111, 107,
		97, 110}

	var count int
	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &count)
	if err != nil {
		t.Errorf("bert decoding error: %s", err)
		return
	}
	if count != 110 {
		t.Errorf("unexpected result: %d (%d)", count, 110)
	}
}

func TestDecodeReplyToNil(t *testing.T) {
	// {reply, {ok, 110}}
	input := []byte{0, 0, 0, 20, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 2, 111, 107,
		97, 110}

	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, nil)
	if err != nil {
		t.Errorf("bert decoding error: %s", err)
		return
	}
}

func TestDecodeOkStruct(t *testing.T) {
	t.Skip("To implement")

	// {reply, {"t1@localhost", "t2@localhost"}}
	input := []byte{0, 0, 0, 47, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 109, 0, 0, 0, 12,
		116, 49, 64, 108, 111, 99, 97, 108, 104, 111, 115, 116, 109, 0, 0, 0, 12, 116,
		50, 64, 108, 111, 99, 97, 108, 104, 111, 115, 116}

	var result struct {
		From string
		To   string
	}
	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &result)
	if err != nil {
		t.Errorf("bert decoding error: %s", err)
		return
	}
}
