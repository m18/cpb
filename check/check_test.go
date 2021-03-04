package check

import (
	"fmt"
	"testing"
)

func TestStringMapsAreEqual(t *testing.T) {
	tests := []struct {
		x  map[string]string
		y  map[string]string
		ok bool
	}{
		{
			x:  nil,
			y:  nil,
			ok: true,
		},
		{
			x:  map[string]string{},
			y:  map[string]string{},
			ok: true,
		},
		{
			x:  map[string]string{"1": "one", "2": "two"},
			y:  map[string]string{"1": "one", "2": "two"},
			ok: true,
		},
		{
			x:  map[string]string{"1": "one", "2": "two"},
			y:  map[string]string{"2": "two", "1": "one"},
			ok: true,
		},
		{
			x:  map[string]string{"1": "one", "2": "two"},
			y:  map[string]string{"1": "one", "2": "one"},
			ok: false,
		},
		{
			x:  map[string]string{"1": "one", "2": "two"},
			y:  map[string]string{"1": "one"},
			ok: false,
		},
		{
			x:  map[string]string{"1": "one", "2": "two"},
			y:  map[string]string{"2": "two"},
			ok: false,
		},
		{
			x:  map[string]string{},
			y:  map[string]string{"1": "one"},
			ok: false,
		},
		{
			x:  map[string]string{"1": "one"},
			y:  map[string]string{},
			ok: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v == %v", test.x, test.y), func(t *testing.T) {
			t.Parallel()
			if StringMapsAreEqual(test.x, test.y) != test.ok {
				t.Errorf("expected %v did not get it", test.ok)
			}
		})
	}
}

func TestStringSlicesAreEqual(t *testing.T) {
	tests := []struct {
		x  []string
		y  []string
		ok bool
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
