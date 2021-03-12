package config

import (
	"strings"
	"testing"
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
			rc, err := newRawConfig().from(strings.NewReader(test.str))
			if err == nil == test.err {
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

func TestRawConfigMerge(t *testing.T) {
	tests := []struct {
		desc      string
		base      func() *rawConfig
		secondary func() *rawConfig
	}{
		{
			desc: "partial intersection",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Port = testExpectedPort
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			secondary: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = "foo"
				res.DB.Host = testExpectedHost
				res.DB.Port = -1
				res.DB.Name = testExpectedName
				return res
			},
		},
		{
			desc: "no intersection",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Port = testExpectedPort
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			secondary: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Host = testExpectedHost
				res.DB.Name = testExpectedName
				return res
			},
		},
		{
			desc: "base only",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			secondary: func() *rawConfig { return newRawConfig() },
		},
		{
			desc: "base only, nil secondary",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			secondary: func() *rawConfig { return nil },
		},
		{
			desc: "secondary only",
			base: func() *rawConfig { return newRawConfig() },
			secondary: func() *rawConfig {
				res := newRawConfig()
				res.Protoc = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			base := test.base()
			secondary := test.secondary()
			base.merge(secondary)
			if err := testRawConfigCheck(base); err != nil {
				t.Error(err)
			}
		})
	}
}
