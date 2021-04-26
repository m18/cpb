package printer

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/m18/cpb/internal/testcheck"
)

func TestTypeOf(t *testing.T) {
	tests := []struct {
		vals     []interface{}
		expected typ
	}{
		{
			vals:     []interface{}{1, 2, int8(3)},
			expected: typeInt,
		},
		{
			vals:     []interface{}{1.1, 2.30},
			expected: typeFloat,
		},
		{
			vals:     []interface{}{true, false},
			expected: typeBool,
		},
		{
			vals:     []interface{}{"foo", "bar"},
			expected: typeString,
		},
		{
			vals:     []interface{}{[]byte{1, 2, 3}, []byte{4, 5, 6, 7, 8}},
			expected: typeBytes,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.vals), func(t *testing.T) {
			t.Parallel()
			for _, v := range test.vals {
				typ := typeOf(v)
				if typ != test.expected {
					t.Fatalf("expected %d but got %d", test.expected, typ)
				}
			}
		})
	}
}

func TestWithFormat(t *testing.T) {
	tests := []struct {
		format format
		err    bool
	}{
		{
			format: FormatTable,
		},
		{
			format: FormatCSV,
		},
		{
			format: FormatTSV,
		},
		{
			format: "foo",
			err:    true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.format), func(t *testing.T) {
			t.Parallel()
			f := WithFormat(test.format)
			b := &formatterBuilder{}

			testcheck.FatalIfUnexpected(t, f(b), test.err)
			if test.err {
				return
			}

			if b.format != test.format {
				t.Fatalf("expected %q but got %q", test.format, b.format)
			}
		})
	}
}

func TestWithHeader(t *testing.T) {
	tests := []struct {
		header bool
	}{
		{
			header: true,
		},
		{
			header: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%t", test.header), func(t *testing.T) {
			t.Parallel()
			f := WithHeader(test.header)
			b := &formatterBuilder{}

			testcheck.FatalIfUnexpected(t, f(b), false)

			if b.header != test.header {
				t.Fatalf("expected %t but did not get it", test.header)
			}
		})
	}
}

func TestWithSpacing(t *testing.T) {
	tests := []struct {
		spacing int
	}{
		{
			spacing: 0,
		},
		{
			spacing: 10,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(strconv.Itoa(test.spacing), func(t *testing.T) {
			t.Parallel()
			f := WithSpacing(test.spacing)
			b := &formatterBuilder{}

			testcheck.FatalIfUnexpected(t, f(b), false)

			if b.spacing != test.spacing {
				t.Fatalf("expected %d but got %d", test.spacing, b.spacing)
			}
		})
	}
}

func TestFormatterBuilderBuild(t *testing.T) {
	checkTable := func(f formatter, header bool) error {
		tf, ok := f.(*tableFormatter)
		if !ok {
			return fmt.Errorf("expected %T but got %T", &tableFormatter{}, f)
		}
		if tf.header != header {
			return fmt.Errorf("expected header to be %t but it was not", header)
		}
		return nil
	}
	tests := []struct {
		format format
		header bool
		check  func(formatter, bool) error
		err    bool
	}{
		{
			format: "",
			header: true,
			check:  checkTable,
		},
		{
			format: FormatTable,
			header: false,
			check:  checkTable,
		},
		{
			format: "foo",
			err:    true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%q, header: %t", test.format, test.header), func(t *testing.T) {
			t.Parallel()
			b := &formatterBuilder{
				format: test.format,
				header: test.header,
			}
			f, err := b.build()
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			testcheck.FatalIf(t, test.check(f, test.header))
		})
	}
}

func TestFormatIsValid(t *testing.T) {
	tests := []struct {
		format format
		err    bool
	}{
		{
			format: FormatTable,
		},
		{
			format: FormatCSV,
		},
		{
			format: FormatTSV,
		},
		{
			format: "foo",
			err:    true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(string(test.format), func(t *testing.T) {
			t.Parallel()
			testcheck.FatalIfUnexpected(t, test.format.isValid(), test.err)
		})
	}
}
