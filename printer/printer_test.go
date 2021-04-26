package printer

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/m18/cpb/internal/testcheck"
)

func TestPrinterNew(t *testing.T) {
	tests := []struct {
		desc    string
		options []func(*formatterBuilder) error
		err     bool
	}{
		{
			desc:    "default (nil)",
			options: nil,
			err:     false,
		},
		{
			desc:    "default (empty)",
			options: []func(*formatterBuilder) error{},
			err:     false,
		},
		{
			desc: "header",
			options: []func(*formatterBuilder) error{
				WithHeader(true),
			},
			err: false,
		},
		{
			desc: "table + header",
			options: []func(*formatterBuilder) error{
				WithFormat(FormatTable),
				WithHeader(false),
			},
			err: false,
		},
		{
			desc: "table + header + spacing",
			options: []func(*formatterBuilder) error{
				WithFormat(FormatTable),
				WithHeader(false),
				WithSpacing(2),
			},
			err: false,
		},
		{
			desc: "invalid format",
			options: []func(*formatterBuilder) error{
				WithFormat("foo"),
			},
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p, err := New(io.Discard, test.options...)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if p == nil {
				t.Fatalf("expected printer to not be nil but it was")
			}
		})
	}
}

func TestPrinterPrint(t *testing.T) {
	cols := []string{"id", "name"}
	rows := [][]interface{}{
		{1, "one"},
		{20, "twenty"},
	}
	expectedHeaderSpacing0 := "idname  \n--------\n"
	expectedTableSpacing0 := " 1one   \n20twenty\n"
	expectedHeaderSpacing2 := "id  name    \n--  ------  \n"
	expectedTableSpacing2 := " 1  one     \n20  twenty  \n"
	tests := []struct {
		desc     string
		options  []func(*formatterBuilder) error
		expected string
	}{
		{
			desc:     "default",
			options:  nil,
			expected: expectedTableSpacing0,
		},
		{
			desc: "header",
			options: []func(*formatterBuilder) error{
				WithHeader(true),
			},
			expected: expectedHeaderSpacing0 + expectedTableSpacing0,
		},
		{
			desc: "spacing",
			options: []func(*formatterBuilder) error{
				WithSpacing(2),
			},
			expected: expectedTableSpacing2,
		},
		{
			desc: "header + spacing",
			options: []func(*formatterBuilder) error{
				WithHeader(true),
				WithSpacing(2),
			},
			expected: expectedHeaderSpacing2 + expectedTableSpacing2,
		},
		{
			desc: "header + spacing + format",
			options: []func(*formatterBuilder) error{
				WithHeader(true),
				WithSpacing(2),
				WithFormat(FormatTable),
			},
			expected: expectedHeaderSpacing2 + expectedTableSpacing2,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			p, err := New(&buf, test.options...)
			testcheck.FatalIf(t, err)
			p.Print(cols, rows)
			res := buf.String()
			if res != test.expected {
				t.Fatalf("expected %q but go %q", test.expected, res)
			}
		})
	}
}

func TestWritefN(t *testing.T) {
	var w writef = func(format string, args ...interface{}) {
		expected := "\n"
		if format != expected {
			t.Fatalf("expected %q but got %q", expected, format)
		}
	}
	w.n()
}

func TestWritefRepeat(t *testing.T) {
	tests := []struct {
		s        string
		count    int
		expected string
	}{
		{
			s:        " ",
			count:    4,
			expected: "    ",
		},
		{
			s:        "",
			count:    4,
			expected: "",
		},
		{
			s:        "12",
			count:    2,
			expected: "1212",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%d x %q", test.count, test.s), func(t *testing.T) {
			t.Parallel()
			var w writef = func(format string, args ...interface{}) {
				if format != test.expected {
					t.Fatalf("expected %q but got %q", test.expected, format)
				}
			}
			w.repeat(test.s, test.count)
		})
	}
}
