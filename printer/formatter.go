package printer

import "fmt"

type format string

func (f format) isValid() error {
	switch f {
	case FormatTable, FormatCSV, FormatTSV:
		return nil
	}
	return fmt.Errorf("invalid format: %s", f)
}

const (
	FormatTable format = "table"
	FormatCSV          = "csv"
	FormatTSV          = "tsv"
)

type typ byte

const (
	typeDefault typ = iota
	typeString
	typeBool
	typeInt
	typeFloat
	typeBytes
)

var verbs = map[typ]byte{
	typeDefault: 'v',
	typeString:  's',
	typeBool:    't',
	typeInt:     'd',
	typeFloat:   'g',
	typeBytes:   'v',
}

type formatter interface {
	format(w writef, cols []string, rows [][]interface{})
}

type formatterBuilder struct {
	format  format
	header  bool
	spacing int
}

func (b *formatterBuilder) build() (formatter, error) {
	var res formatter
	switch b.format {
	case "", FormatTable:
		res = newTableFormatter(b.header, b.spacing)
	// TODO:
	// case CSV:
	// case TSV:
	default:
		return nil, fmt.Errorf("unknown format: %q", b.format)
	}
	return res, nil
}

func WithFormat(f format) func(*formatterBuilder) error {
	return func(b *formatterBuilder) error {
		if err := f.isValid(); err != nil {
			return err
		}
		b.format = f
		return nil
	}
}

func WithHeader(yes bool) func(*formatterBuilder) error {
	return func(b *formatterBuilder) error {
		b.header = yes
		return nil
	}
}

func WithSpacing(s int) func(*formatterBuilder) error {
	return func(b *formatterBuilder) error {
		b.spacing = s
		return nil
	}
}

type column struct {
	name         string
	headerFormat string
	format       string
	width        int
}

func typeOf(v interface{}) typ {
	switch v.(type) {
	// byte is uint8
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return typeInt
	case float32, float64:
		return typeFloat
	case string:
		return typeString
	case []byte:
		return typeBytes
	case bool:
		return typeBool
	}
	return typeDefault
}
