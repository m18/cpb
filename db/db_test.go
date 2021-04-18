package db

import (
	"fmt"
	"testing"

	"github.com/m18/cpb/check"
)

// TODO: test the rest with a contenerized DB

func TestBtoi(t *testing.T) {
	tests := []struct {
		desc  string
		input [][]byte
	}{
		{
			desc:  "nil input",
			input: nil,
		},
		{
			desc:  "empty input",
			input: [][]byte{},
		},
		{
			desc:  "non-empty slice",
			input: [][]byte{{1}, {2}, {3}},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			res := btoi(test.input)
			if (test.input == nil) != (res == nil) {
				var pre, post string
				if test.input != nil {
					pre = "not "
				} else {
					post = " not"
				}
				t.Fatalf("expected result to %sbe nil but it was%s", pre, post)
			}
			if len(res) != len(test.input) {
				t.Fatalf("expected result to have lenght %d but it was %d", len(test.input), len(res))
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	const stringerRes = "ok"
	stringer := func([]byte) (string, error) {
		return stringerRes, nil
	}
	tests := []struct {
		dbVal       interface{}
		useStringer bool
		expected    interface{}
	}{
		{
			dbVal:       1,
			useStringer: true,
			expected:    1,
		},
		{
			dbVal:       1,
			useStringer: false,
			expected:    1,
		},
		{
			dbVal:       []byte{},
			useStringer: true,
			expected:    stringerRes,
		},
		{
			dbVal:       []byte{},
			useStringer: false,
			expected:    []byte{},
		},
		{
			dbVal:       []byte{1, 2, 3},
			useStringer: false,
			expected:    []byte{1, 2, 3},
		},
		{
			dbVal:       nil,
			useStringer: true,
			expected:    nil,
		},
		{
			dbVal:       nil,
			useStringer: false,
			expected:    nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("dbVal: %v. stringer: %t", test.dbVal, test.useStringer), func(t *testing.T) {
			t.Parallel()
			var s func([]byte) (string, error)
			if test.useStringer {
				s = stringer
			}
			res, _ := getValue(test.dbVal, s)
			if b, ok := test.expected.([]byte); ok {
				if !check.ByteSlicesAreEqual(b, res.([]byte)) {
					t.Fatalf("expected %v but got %v", test.expected, res)
				}
			} else if res != test.expected {
				t.Fatalf("expected %v but got %v", test.expected, res)
			}
		})
	}
}
