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

func TestStringToSimpleTypeMapsAreEqual(t *testing.T) {
	tests := []struct {
		x  map[string]interface{}
		y  map[string]interface{}
		ok bool
	}{
		{
			x:  nil,
			y:  nil,
			ok: true,
		},
		{
			x:  map[string]interface{}{},
			y:  map[string]interface{}{},
			ok: true,
		},
		{
			x:  map[string]interface{}{"1": "one", "2": "two"},
			y:  map[string]interface{}{"1": "one", "2": "two"},
			ok: true,
		},
		{
			x:  map[string]interface{}{"1": "one", "2": "two"},
			y:  map[string]interface{}{"2": "two", "1": "one"},
			ok: true,
		},
		{
			x:  map[string]interface{}{"1": 1, "2": 2},
			y:  map[string]interface{}{"2": 2, "1": 1},
			ok: true,
		},
		{
			x:  map[string]interface{}{"1": "one", "2": "two"},
			y:  map[string]interface{}{"1": "one", "2": "one"},
			ok: false,
		},
		{
			x:  map[string]interface{}{"1": "one", "2": "two"},
			y:  map[string]interface{}{"1": "one"},
			ok: false,
		},
		{
			x:  map[string]interface{}{"1": "one", "2": "two"},
			y:  map[string]interface{}{"2": "two"},
			ok: false,
		},
		{
			x:  map[string]interface{}{},
			y:  map[string]interface{}{"1": "one"},
			ok: false,
		},
		{
			x:  map[string]interface{}{"1": "one"},
			y:  map[string]interface{}{},
			ok: false,
		},
		{
			x:  map[string]interface{}{"1": "1"},
			y:  map[string]interface{}{"1": 1},
			ok: false,
		},
		{
			x:  map[string]interface{}{"1": 1},
			y:  map[string]interface{}{"1": "1"},
			ok: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v == %v", test.x, test.y), func(t *testing.T) {
			t.Parallel()
			if StringToSimpleTypeMapsAreEqual(test.x, test.y) != test.ok {
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

func TestStringSetsAreEqual(t *testing.T) {
	tests := []struct {
		x  map[string]struct{}
		y  map[string]struct{}
		ok bool
	}{
		{
			x:  nil,
			y:  nil,
			ok: true,
		},
		{
			x:  map[string]struct{}{},
			y:  map[string]struct{}{},
			ok: true,
		},
		{
			x:  map[string]struct{}{"1": {}, "2": {}},
			y:  map[string]struct{}{"1": {}, "2": {}},
			ok: true,
		},
		{
			x:  map[string]struct{}{"1": {}, "2": {}},
			y:  map[string]struct{}{"2": {}, "1": {}},
			ok: true,
		},
		{
			x:  map[string]struct{}{"1": {}, "2": {}},
			y:  map[string]struct{}{"1": {}},
			ok: false,
		},
		{
			x:  map[string]struct{}{"1": {}, "2": {}},
			y:  map[string]struct{}{"2": {}},
			ok: false,
		},
		{
			x:  map[string]struct{}{},
			y:  map[string]struct{}{"1": {}},
			ok: false,
		},
		{
			x:  map[string]struct{}{"1": {}},
			y:  map[string]struct{}{},
			ok: false,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v == %v", test.x, test.y), func(t *testing.T) {
			t.Parallel()
			if StringSetsAreEqual(test.x, test.y) != test.ok {
				t.Errorf("expected %v did not get it", test.ok)
			}
		})
	}
}
