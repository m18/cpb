package printer

import (
	"fmt"
	"io"
	"strings"
)

type writef func(string, ...interface{})

func (w writef) n() {
	w("\n")
}

func (w writef) repeat(s string, count int) {
	w(strings.Repeat(s, count))
}

type Printer struct {
	w writef
	f formatter
}

func New(w io.Writer, options ...func(*formatterBuilder) error) (*Printer, error) {
	b := &formatterBuilder{}
	for _, option := range options {
		if err := option(b); err != nil {
			return nil, err
		}
	}
	f, err := b.build()
	if err != nil {
		return nil, err
	}
	return &Printer{
		w: func(format string, args ...interface{}) {
			fmt.Fprintf(w, format, args...)
		},
		f: f,
	}, nil
}

// TODO: context
func (p *Printer) Print(cols []string, rows [][]interface{}) {
	p.f.format(p.w, cols, rows)
}
