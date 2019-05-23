package bert_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/processone/bert"
)

// Small Erlang Term type is Uint8. It cannot fit into an int8
func TestDecodeInt8(t *testing.T) {
	var i int8
	buf := bytes.NewBuffer([]byte{131, 97, 255})
	if err := bert.Decode(buf, &i); err != bert.ErrRange {
		t.Errorf("Decoding an Erlang small integer into int8 should fail")
	}
}

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		input []byte
		want  int64
	}{
		{input: []byte{131, 97, 42}, want: 42},
		{input: []byte{131, 97, 255}, want: 255},
		{input: []byte{131, 98, 255, 255, 255, 0}, want: -256},
		{input: []byte{131, 98, 0, 0, 1, 0}, want: 256},
		{input: []byte{131, 98, 128, 0, 0, 0}, want: -2147483648},
		{input: []byte{131, 98, 127, 255, 255, 255}, want: 2147483647},
	}

	for _, tc := range tests {
		var i int
		buf := bytes.NewBuffer(tc.input)
		if err := bert.Decode(buf, &i); err != nil {
			t.Errorf("cannot decode Erlang term: %s", err)
			return
		}

		if int64(i) != tc.want {
			t.Errorf("incorrect decoded value: %d. expected: %d", i, tc.want)
		}
	}
}

// TODO: Implement decode same types to []byte and bert.Atom
func TestDecodeToString(t *testing.T) {
	longUTF8 := strings.Repeat("🖖", 64)
	tests := []struct {
		input []byte
		want  string
	}{
		{input: []byte{131, 100, 0, 0}, want: ""},
		{input: []byte{131, 100, 0, 2, 111, 107}, want: "ok"},
		{input: []byte{131, 119, 4, 240, 159, 150, 150}, want: "🖖"},
		{input: append([]byte{131, 118, 1, 0}, []byte(longUTF8)...), want: longUTF8},
		{input: []byte{131, 107, 0, 5, 72, 101, 108, 108, 111}, want: "Hello"},
		{input: []byte{131, 109, 0, 0, 0, 5, 72, 101, 108, 108, 111}, want: "Hello"},
		{input: []byte{131, 109, 0, 0, 0, 10, 240, 159, 150, 150, 32, 72, 101, 108, 108, 111}, want: "🖖 Hello"},
		{input: []byte{131, 108, 0, 0, 0, 3, 98, 0, 1, 245, 150, 97, 72, 97, 105, 106}, want: "🖖Hi"},
	}

	for _, tc := range tests {
		var a string
		buf := bytes.NewBuffer(tc.input)
		if err := bert.Decode(buf, &a); err != nil {
			t.Errorf("cannot decode Erlang term: %s", err)
			return
		}

		if a != tc.want {
			t.Errorf("incorrect decoded value: %#v. expected: %#v", a, tc.want)
		}
	}
}

func TestDecodeEmptyTuple(t *testing.T) {
	input := []byte{131, 104, 0}
	want := struct{}{}

	var tuple struct{}
	buf := bytes.NewBuffer(input)
	if err := bert.Decode(buf, &tuple); err != nil {
		t.Errorf("cannot decode Erlang term: %s", err)
		return
	}

	if tuple != want {
		t.Errorf("cannot decode empty tuple: %v", tuple)
	}
}

// Decode a tuple with two elements.
func TestDecodeTuple2(t *testing.T) {
	input := []byte{131, 104, 2, 100, 0, 5, 101, 114, 114, 111, 114, 100, 0, 9, 110, 111,
		116, 95, 102, 111, 117, 110, 100}
	want := struct {
		Result string
		Reason string
	}{"error", "not_found"}

	var tuple struct {
		Result string
		Reason string
	}
	buf := bytes.NewBuffer(input)
	if err := bert.Decode(buf, &tuple); err != nil {
		t.Errorf("cannot decode Erlang term: %s", err)
		return
	}

	if tuple != want {
		t.Errorf("cannot decode empty tuple: %v", tuple)
	}
}

func TestFailOnLengthMismatch(t *testing.T) {
	input := []byte{131, 104, 2, 100, 0, 5, 101, 114, 114, 111, 114, 100, 0, 9, 110, 111,
		116, 95, 102, 111, 117, 110, 100}

	var tuple struct {
		Result string
		Reason string
		Extra  string
	}
	buf := bytes.NewBuffer(input)
	if err := bert.Decode(buf, &tuple); err == nil {
		t.Errorf("decoding tuple into struct with different number of field should fail")
	}
}

type result1 struct {
	Tag    string `erlang:"tag"`
	Result string `erlang:"tag:ok"`
	Reason string `erlang:"tag:error"`
}

func TestDecodeResult(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  result1
	}{
		// Erlang function returns:
		{name: "ok", input: []byte{131, 100, 0, 2, 111, 107}, want: result1{Tag: "ok"}},
		{name: "error", input: []byte{131, 100, 0, 5, 101, 114, 114, 111, 114}, want: result1{Tag: "error"}},
		{name: "info", input: []byte{131, 100, 0, 4, 105, 110, 102, 111}, want: result1{Tag: "info"}},
		{name: "{ok, Result}", input: []byte{131, 104, 2, 100, 0, 2, 111, 107, 100, 0, 5, 102, 111, 117, 110, 100},
			want: result1{Tag: "ok", Result: "found"}},
		{name: "{error, Reason}", input: []byte{131, 104, 2, 100, 0, 5, 101, 114, 114, 111, 114, 100, 0, 9, 110, 111, 116, 95,
			102, 111, 117, 110, 100}, want: result1{Tag: "error", Reason: "not_found"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(st *testing.T) {
			var res result1
			buf := bytes.NewBuffer(tc.input)

			if err := bert.Decode(buf, &res); err != nil {
				st.Errorf("cannot decode function call result: %s", err)
				return
			}

			if tc.want.Tag != res.Tag {
				st.Errorf("incorrect Tag: %v (!= %v)", res.Tag, tc.want.Tag)
			}
			if tc.want.Result != res.Result {
				st.Errorf("incorrect Result: %v (!= %v)", res.Result, tc.want.Result)
			}
			if tc.want.Reason != res.Reason {
				st.Errorf("incorrect Reason: %v (!= %v)", res.Reason, tc.want.Reason)
			}
		})
	}
}

func TestDecodeTupleResult(t *testing.T) {
	input := []byte{131, 104, 4, 97, 1, 97, 2, 97, 3, 97, 4}
	want := struct {
		A int
		B int
		C int
		D int
	}{1, 2, 3, 4}

	var tuple struct {
		A int
		B int
		C int
		D int
	}
	buf := bytes.NewBuffer(input)

	if err := bert.Decode(buf, &tuple); err != nil {
		t.Errorf("cannot decode Erlang term: %s", err)
		return
	}

	if tuple != want {
		t.Errorf("result does not match expectation: %v", tuple)
	}
}

// We have a bert.String type that allow developer to know if the return struct was an atom when this matters.
// For example, it can be use to make a difference between the atom result not_found and the value "not_found".
func TestDecodeAtomVsString(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  bert.String
	}{
		{name: "false as atom", input: []byte{131, 100, 0, 5, 102, 97, 108, 115, 101}, want: bert.A("false")},
		{name: "false as result", input: []byte{131, 107, 0, 5, 102, 97, 108, 115, 101}, want: bert.S("false")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(st *testing.T) {
			var res bert.String
			buf := bytes.NewBuffer(tc.input)

			if err := bert.Decode(buf, &res); err != nil {
				st.Errorf("cannot decode function call result: %s", err)
				return
			}

			if tc.want != res {
				st.Errorf("incorrect result: %#v (!= %#v)", res, tc.want)
			}
		})
	}
}
