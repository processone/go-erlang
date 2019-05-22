package bert_test

import (
	"bytes"
	"testing"

	"github.com/processone/bert"
)

// TODO: Refactor the test to work with both the Erlang raw term format and the Bert reply packet
func TestDecodeSimple(t *testing.T) {
	// {reply, valid}
	input := []byte{0, 0, 0, 19, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 100, 0, 5, 118, 97,
		108, 105, 100}

	var result struct {
		Value string `erlang:"tag"`
	}

	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &result)
	if err != nil {
		t.Errorf("bert decoding failed: %s", err)
		return
	}
	if result.Value != "valid" {
		t.Errorf("unexpected result: %s", result.Value)
	}
}

func TestDecodeErrorReply(t *testing.T) {
	// {reply, {error, exists}}
	input := []byte{0, 0, 0, 30, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 5, 101, 114, 114, 111, 114, 100,
		0, 6, 101, 120, 105, 115, 116, 115}

	var result struct {
		Tag    string `erlang:"tag"`
		Reason string `erlang:"tag:error"`
		Result string `erlang:"tag:ok"`
	}
	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &result)
	if err != nil {
		t.Errorf("bert decoding failed: %s", err)
		return
	}

	if result.Tag != "error" {
		t.Errorf("unexpected tag value: '%s'", result.Tag)
	}
	if result.Reason != "exists" {
		t.Errorf("unexpected error reason: '%s'", result.Reason)
	}
	if result.Result != "" {
		t.Errorf("result is expected to be empty: '%s'", result.Result)
	}
}

func TestDecodeOkReply(t *testing.T) {
	// {reply, {ok, 110}}
	input := []byte{0, 0, 0, 20, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 2, 111, 107,
		97, 110}

	var result struct {
		Tag    string `erlang:"tag"`
		Reason string `erlang:"tag:error"`
		Count  int    `erlang:"tag:ok"`
	}
	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, &result)
	if err != nil {
		t.Errorf("bert decoding failed: %s", err)
		return
	}
	if result.Count != 110 {
		t.Errorf("unexpected count value: %d (%d)", result.Count, 110)
	}
	if result.Reason != "" {
		t.Errorf("reason is expected to be empty: '%s'", result.Reason)
	}
}

func TestDecodeReplyToNil(t *testing.T) {
	// {reply, {ok, 110}}
	input := []byte{0, 0, 0, 20, 131, 104, 2, 100, 0, 5, 114, 101, 112, 108, 121, 104, 2, 100, 0, 2, 111, 107,
		97, 110}

	buf := bytes.NewBuffer(input)
	err := bert.DecodeReply(buf, nil)
	if err == nil {
		t.Errorf("bert decoding to nil should fail")
	}
}

func TestDecodeOkStruct(t *testing.T) {
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
		t.Errorf("bert decoding failed: %s", err)
		return
	}

	if result.From != "t1@localhost" {
		t.Errorf("incorrect from: %s", result.From)
	}
	if result.To != "t2@localhost" {
		t.Errorf("incorrect from: %s", result.To)
	}
}
