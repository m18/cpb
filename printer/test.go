package printer

import (
	"bytes"
	"fmt"
	"testing"
)

func makeTestWritef(t *testing.T) (writef, func(string)) {
	var buf bytes.Buffer
	writef := func(format string, args ...interface{}) {
		buf.WriteString(fmt.Sprintf(format, args...))
	}
	checkWrote := func(expected string) {
		actual := buf.String()
		if actual != expected {
			t.Fatalf("expected %q but got %q", expected, actual)
		}
	}
	return writef, checkWrote
}
