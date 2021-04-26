package printer

import (
	"fmt"
	"testing"
)

func TestTableFormatterAlign(t *testing.T) {
	tests := []struct {
		desc     string
		typ      typ
		expected alignment
	}{
		{
			desc:     "default",
			typ:      typeDefault,
			expected: alignLeft,
		},
		{
			desc:     "string",
			typ:      typeString,
			expected: alignLeft,
		},
		{
			desc:     "bool",
			typ:      typeBool,
			expected: alignLeft,
		},
		{
			desc:     "int",
			typ:      typeInt,
			expected: alignRight,
		},
		{
			desc:     "float",
			typ:      typeFloat,
			expected: alignRight,
		},
		{
			desc:     "bytes",
			typ:      typeBytes,
			expected: alignLeft,
		},
		{
			desc:     "string",
			typ:      typeString,
			expected: alignLeft,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			f := &tableFormatter{}
			res := f.align(test.typ)
			if res != test.expected {
				t.Fatalf("expected %d but got %d", test.expected, res)
			}
		})
	}
}

func TestTableFormatterWriteCellSpacing(t *testing.T) {
	tests := []struct {
		cellSpacing string
		expected    string
	}{
		{
			cellSpacing: "",
			expected:    "",
		},
		{
			cellSpacing: " ",
			expected:    " ",
		},
		{
			cellSpacing: "  ",
			expected:    "  ",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%q", test.cellSpacing), func(t *testing.T) {
			t.Parallel()
			w, checkWrote := makeTestWritef(t)
			f := &tableFormatter{cellSpacing: test.cellSpacing}
			f.writeCellSpacing(w)
			checkWrote(test.expected)
		})
	}
}

func TestTableFormatterWriteCell(t *testing.T) {
	tests := []struct {
		desc        string
		cellSpacing string
		format      string
		val         interface{}
		expected    string
	}{
		{
			desc:        "bare string",
			cellSpacing: "",
			format:      "%" + string(verbs[typeString]),
			val:         "foo",
			expected:    "foo",
		},
		{
			desc:        "string + spacing",
			cellSpacing: " ",
			format:      "%" + string(verbs[typeString]),
			val:         "foo",
			expected:    "foo ",
		},
		{
			desc:        "string + spacing + left padding",
			cellSpacing: " ",
			format:      "%4" + string(verbs[typeString]),
			val:         "foo",
			expected:    " foo ",
		},
		{
			desc:        "string + spacing + right padding",
			cellSpacing: " ",
			format:      "%-4" + string(verbs[typeString]),
			val:         "foo",
			expected:    "foo  ",
		},
		{
			desc:        "int",
			cellSpacing: "",
			format:      "%2" + string(verbs[typeInt]),
			val:         1,
			expected:    " 1",
		},
		{
			desc:        "float",
			cellSpacing: "",
			format:      "%-5" + string(verbs[typeFloat]),
			val:         1.23,
			expected:    "1.23 ",
		},
		{
			desc:        "float, trailing zeros",
			cellSpacing: "",
			format:      "%5" + string(verbs[typeFloat]),
			val:         1.23000,
			expected:    " 1.23",
		},
		{
			desc:        "bytes",
			cellSpacing: "",
			format:      "%0" + string(verbs[typeBytes]),
			val:         []byte{1, 2, 3},
			expected:    "[1 2 3]",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			w, checkWrote := makeTestWritef(t)
			f := &tableFormatter{cellSpacing: test.cellSpacing}
			f.writeCell(w, test.format, test.val)
			checkWrote(test.expected)
		})
	}
}

func TestTableFormatterWriteRows(t *testing.T) {
	cols := []*column{
		{format: "%-5s"},
		{format: "%3d"},
	}
	tests := []struct {
		desc     string
		cols     []*column
		rows     [][]interface{}
		expected string
	}{
		{
			desc:     "0 cols, 0 row",
			cols:     []*column{},
			rows:     [][]interface{}{},
			expected: "",
		},
		{
			desc:     "1 cols, 0 row",
			cols:     []*column{},
			rows:     [][]interface{}{},
			expected: "",
		},
		{
			desc: "1 cols, 1 row",
			cols: cols[:1],
			rows: [][]interface{}{
				{"one"},
			},
			expected: "one  \n",
		},
		{
			desc: "1 cols, 2 row",
			cols: cols[:1],
			rows: [][]interface{}{
				{"one"},
				{"two"},
			},
			expected: "one  \ntwo  \n",
		},
		{
			desc: "2 cols, 1 row",
			cols: cols,
			rows: [][]interface{}{
				{"one", 1},
			},
			expected: "one    1\n",
		},
		{
			desc: "2 cols, 2 row",
			cols: cols,
			rows: [][]interface{}{
				{"one", 1},
				{"two", 2},
			},
			expected: "one    1\ntwo    2\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			w, checkWrote := makeTestWritef(t)
			f := &tableFormatter{}
			f.writeRows(w, test.cols, test.rows)
			checkWrote(test.expected)
		})
	}
}

func TestTableFormatterWriteHeader(t *testing.T) {
	tests := []struct {
		desc     string
		cols     []*column
		expected string
	}{
		{
			desc: "1 col, empty name",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: ""},
			},
			expected: "     \n-----\n",
		},
		{
			desc: "1 col, shorter name",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: "foo"},
			},
			expected: "foo  \n-----\n",
		},
		{
			desc: "1 col, full-length name",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: "foo01"},
			},
			expected: "foo01\n-----\n",
		},
		{
			desc: "2 cols, empty names",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: ""},
				{headerFormat: "%-3s", width: 3, name: ""},
			},
			expected: "        \n--------\n",
		},
		{
			desc: "2 cols, 1 shorter name, 1 full-length name",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: "foo"},
				{headerFormat: "%-3s", width: 3, name: "bar"},
			},
			expected: "foo  bar\n--------\n",
		},
		{
			desc: "2 cols, 2 full-length names",
			cols: []*column{
				{headerFormat: "%-5s", width: 5, name: "foo01"},
				{headerFormat: "%-3s", width: 3, name: "bar"},
			},
			expected: "foo01bar\n--------\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			w, checkWrote := makeTestWritef(t)
			f := &tableFormatter{}
			f.writeHeader(w, test.cols)
			checkWrote(test.expected)
		})
	}
}

func TestTableFormatterColumns(t *testing.T) {
	tests := []struct {
		desc     string
		cols     []string
		rows     [][]interface{}
		expected []*column
	}{
		{
			desc:     "empty cols and rows",
			cols:     []string{},
			rows:     [][]interface{}{},
			expected: []*column{},
		},
		{
			desc: "1 col, empty rows",
			cols: []string{"foo"},
			rows: [][]interface{}{},
			expected: []*column{
				{name: "foo", headerFormat: "%-3s", format: "%-3v", width: 3},
			},
		},
		{
			desc: "1 int col, 1 row",
			cols: []string{"foo"},
			rows: [][]interface{}{{12345}},
			expected: []*column{
				{name: "foo", headerFormat: "%-5s", format: "%5d", width: 5},
			},
		},
		{
			desc: "1 float col, 2 row",
			cols: []string{"foo"},
			rows: [][]interface{}{{12.345}, {123.4567}},
			expected: []*column{
				{name: "foo", headerFormat: "%-8s", format: "%8g", width: 8},
			},
		},
		{
			desc: "2 col, 3 row",
			cols: []string{"foo", "bar", "baz"},
			rows: [][]interface{}{
				{[]byte{1, 2, 3}, "one", true},
				{[]byte{4, 5, 6, 7, 8}, "two", false},
				{[]byte{4}, "three", false},
			},
			expected: []*column{
				{name: "foo", headerFormat: "%-11s", format: "%-0v", width: 11},
				{name: "bar", headerFormat: "%-5s", format: "%-5s", width: 5},
				{name: "baz", headerFormat: "%-5s", format: "%-5t", width: 5},
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			f := &tableFormatter{}
			res := f.columns(test.cols, test.rows)
			if len(res) != len(test.expected) {
				t.Fatalf("expected column count to be %d but it was %d", len(test.expected), len(res))
			}
			for i, r := range res {
				e := test.expected[i]
				if r.name != e.name {
					t.Fatalf("expected name to be %q but it was %q", e.name, r.name)
				}
				if r.headerFormat != e.headerFormat {
					t.Fatalf("expected header format to be %q but it was %q", e.headerFormat, r.headerFormat)
				}
				if r.format != e.format {
					t.Fatalf("expected format to be %q but it was %q", e.format, r.format)
				}
				if r.width != e.width {
					t.Fatalf("expected width to be %d but it was %d", e.width, r.width)
				}
			}
		})
	}
}

func TestTableFormatterFormat(t *testing.T) {
	cols := []string{"id", "name"}
	rows := [][]interface{}{
		{1, "one"},
		{20, "twenty"},
	}
	tests := []struct {
		desc     string
		cols     []string
		header   bool
		expected string
	}{
		{
			desc:     "empty cols",
			cols:     []string{},
			header:   true,
			expected: "",
		},
		{
			desc:     "2 cols with header",
			cols:     cols,
			header:   true,
			expected: "idname  \n--------\n 1one   \n20twenty\n",
		},
		{
			desc:     "2 cols w/out header",
			cols:     cols,
			header:   false,
			expected: " 1one   \n20twenty\n",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			w, checkWrote := makeTestWritef(t)
			f := &tableFormatter{header: test.header}
			f.format(w, test.cols, rows)
			checkWrote(test.expected)
		})
	}
}

func TestNewTableFormatter(t *testing.T) {
	tests := []struct {
		header              bool
		cellSpacing         int
		expectedCellSpacing string
	}{
		{
			expectedCellSpacing: "",
		},
		{
			header:              true,
			cellSpacing:         2,
			expectedCellSpacing: "  ",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%q", test.cellSpacing), func(t *testing.T) {
			t.Parallel()
			f := newTableFormatter(test.header, test.cellSpacing)
			if f.header != test.header {
				t.Fatalf("expected header to be %t but is was not", test.header)
			}
			if f.cellSpacing != test.expectedCellSpacing {
				t.Fatalf("expected cell spacing to be %s but is was %s", test.expectedCellSpacing, f.cellSpacing)
			}
		})
	}
}
