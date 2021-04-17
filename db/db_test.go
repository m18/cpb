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
	const printRes = "ok"
	printer := func([]byte) (string, error) {
		return printRes, nil
	}
	tests := []struct {
		dbVal      interface{}
		usePrinter bool
		expected   interface{}
	}{
		{
			dbVal:      1,
			usePrinter: true,
			expected:   1,
		},
		{
			dbVal:      1,
			usePrinter: false,
			expected:   1,
		},
		{
			dbVal:      []byte{},
			usePrinter: true,
			expected:   printRes,
		},
		{
			dbVal:      []byte{},
			usePrinter: false,
			expected:   []byte{},
		},
		{
			dbVal:      []byte{1, 2, 3},
			usePrinter: false,
			expected:   []byte{1, 2, 3},
		},
		{
			dbVal:      nil,
			usePrinter: true,
			expected:   nil,
		},
		{
			dbVal:      nil,
			usePrinter: false,
			expected:   nil,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("dbVal: %v. printer: %t", test.dbVal, test.usePrinter), func(t *testing.T) {
			t.Parallel()
			var p func([]byte) (string, error)
			if test.usePrinter {
				p = printer
			}
			res, _ := getValue(test.dbVal, p)
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
