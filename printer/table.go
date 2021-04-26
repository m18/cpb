package printer

import (
	"fmt"
	"strings"
)

type alignment byte

const (
	alignLeft alignment = iota
	alignRight
)

var alignFlags = map[alignment]string{
	alignLeft: "-",
}

type tableFormatter struct {
	header      bool
	cellSpacing string
}

func newTableFormatter(header bool, cellSpacing int) *tableFormatter {
	return &tableFormatter{
		header:      header,
		cellSpacing: strings.Repeat(" ", cellSpacing),
	}
}

func (f *tableFormatter) format(w writef, cols []string, rows [][]interface{}) {
	if len(cols) == 0 {
		return
	}
	tcols := f.columns(cols, rows)
	if f.header {
		f.writeHeader(w, tcols)
	}
	f.writeRows(w, tcols, rows)
}

// columns creates column metadata.
func (f *tableFormatter) columns(cols []string, rows [][]interface{}) []*column {
	meta := make([]*struct {
		typ   typ
		width int
	}, 0, len(cols))

	for i, col := range cols {
		var val interface{}
		if len(rows) > 0 {
			val = rows[0][i]
		}
		meta = append(meta, &struct {
			typ   typ
			width int
		}{
			typ:   typeOf(val),
			width: len(col),
		})
	}

	// increase col width to fit in its widest value
	for _, row := range rows {
		for i, val := range row {
			var width int
			m := meta[i]
			switch m.typ {
			case typeString:
				width = len(val.(string))
			default:
				format := "%" + string(verbs[m.typ])
				width = len(fmt.Sprintf(format, val))
			}
			if m.width < width {
				m.width = width
			}
		}
	}

	res := make([]*column, 0, len(cols))
	format := "%%%s%d%c"
	for i, col := range cols {
		m := meta[i]
		width := m.width
		if m.typ == typeBytes {
			// do not pad each byte inside slice
			width = 0
		}
		res = append(res, &column{
			name:         col,
			headerFormat: fmt.Sprintf(format, alignFlags[alignLeft], m.width, verbs[typeString]),
			format:       fmt.Sprintf(format, alignFlags[f.align(m.typ)], width, verbs[m.typ]),
			width:        m.width,
		})
	}

	return res
}

func (f *tableFormatter) writeHeader(w writef, cols []*column) {
	for _, col := range cols {
		f.writeCell(w, col.headerFormat, col.name)
	}
	w.n()
	for _, col := range cols {
		w.repeat("-", col.width)
		f.writeCellSpacing(w)
	}
	w.n()
}

func (f *tableFormatter) writeRows(w writef, cols []*column, rows [][]interface{}) {
	for _, row := range rows {
		for i, val := range row {
			col := cols[i]
			f.writeCell(w, col.format, val)
		}
		w.n()
	}
}

func (f *tableFormatter) writeCell(w writef, format string, val interface{}) {
	w(format, val)
	f.writeCellSpacing(w)
}

func (f *tableFormatter) writeCellSpacing(w writef) {
	w(f.cellSpacing)
}

func (f *tableFormatter) align(typ typ) alignment {
	switch typ {
	case typeInt, typeFloat:
		return alignRight
	default:
		return alignLeft
	}
}
