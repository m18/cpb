package check

import (
	"fmt"
	"testing"
)

func TestStringSlicesAreEqual(t *testing.T) {
	tests := []struct {
		x, y []string
		ok   bool
	}{
		{
			x:  nil,
			y:  nil,
			ok: true,
		},
		{
			x:  []string{},
			y:  []string{},
			ok: true,
		},
		{
			x:  []string{"one", "two"},
			y:  []string{"one", "two"},
			ok: true,
		},
		{
			x:  []string{"one", "two"},
			y:  []string{"two", "one"},
			ok: false,
		},
		{
			x:  []string{"one", "two"},
			y:  []string{"one"},
			ok: false,
		},
		{
			x:  []string{"one", "two"},
			y:  []string{"two"},
			ok: false,
		},
		{
			x:  []string{},
			y:  []string{"one"},
			ok: false,
		},
		{
			x:  []string{"one"},
			y:  []string{},
			ok: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v == %v", test.x, test.y), func(t *testing.T) {
			t.Parallel()
			if StringSlicesAreEqual(test.x, test.y) != test.ok {
				t.Errorf("expected %v did not get it", test.ok)
			}
		})
	}
}

func TestByteSlicesAreEqual(t *testing.T) {
	tests := []struct {
		x, y []byte
		ok   bool
	}{
		{
			x:  nil,
			y:  nil,
			ok: true,
		},
		{
			x:  []byte{},
			y:  []byte{},
			ok: true,
		},
		{
			x:  []byte{1, 2},
			y:  []byte{1, 2},
			ok: true,
		},
		{
			x:  []byte{1, 2},
			y:  []byte{2, 1},
			ok: false,
		},
		{
			x:  []byte{1, 2},
			y:  []byte{1},
			ok: false,
		},
		{
			x:  []byte{1, 2},
			y:  []byte{2},
			ok: false,
		},
		{
			x:  []byte{},
			y:  []byte{1},
			ok: false,
		},
		{
			x:  []byte{1},
			y:  []byte{},
			ok: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v == %v", test.x, test.y), func(t *testing.T) {
			t.Parallel()
			if ByteSlicesAreEqual(test.x, test.y) != test.ok {
				t.Errorf("expected %v did not get it", test.ok)
			}
		})
	}
}
