package config

import (
	"testing"

	"github.com/m18/cpb/config/internal"
)

func TestRawConfigFrom(t *testing.T) {
	tests := []struct {
		str string
		err bool
	}{
		{
			str: `{"messages":{"in": {},"out": {}}}`,
			err: false,
		},
		{
			str: `{"messages":{}}`,
			err: false,
		},
		{
			str: `{"messages":{}}`,
			err: false,
		},
		{
			str: `{"foo": "bar"}`,
			err: false,
		},
		{
			str: `{"messages"}`,
			err: true,
		},
		{
			str: ``,
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.str, func(t *testing.T) {
			t.Parallel()
			rc := &rawConfig{}
			if err := rc.from(internal.MakeTestConfigFS(test.str)); err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if rc.DB == nil {
				t.Fatalf("expected DB not to be nil but it was")
			}
			if rc.Messages == nil {
				t.Fatalf("expected Messages not to be nil but it was")
			}
		})
	}
}
