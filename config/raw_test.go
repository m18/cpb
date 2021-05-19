package config

import (
	"testing"

	"github.com/m18/cpb/internal/testcheck"
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
			rc := newRawConfig()
			err := rc.from([]byte(test.str))
			testcheck.FatalIfUnexpected(t, err, test.err)
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
		desc     string
		base     func() *rawConfig
		override func() *rawConfig
		isSet    func(string) bool
	}{
		{
			desc: "partial intersection",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = "foo"
				res.DB.Driver = testExpectedDriver
				res.DB.Port = -1
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			override: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = testExpectedProtoc
				res.Proto.Dir = testExpectedProtoDir
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				return res
			},
			isSet: func(name string) bool {
				switch name {
				case flagProtoc, flagProtoDir, flagHost, flagPort, flagName:
					return true
				default:
					return false
				}
			},
		},
		{
			desc: "no intersection",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Proto.Dir = testExpectedProtoDir
				res.DB.Driver = testExpectedDriver
				res.DB.Port = testExpectedPort
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			override: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = testExpectedProtoc
				res.DB.Host = testExpectedHost
				res.DB.Name = testExpectedName
				return res
			},
			isSet: func(name string) bool {
				switch name {
				case flagProtoc, flagHost, flagName:
					return true
				default:
					return false
				}
			},
		},
		{
			desc: "base only",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = testExpectedProtoc
				res.Proto.Dir = testExpectedProtoDir
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			override: func() *rawConfig { return newRawConfig() },
			isSet:    func(string) bool { return false },
		},
		{
			desc: "override only",
			base: func() *rawConfig { return newRawConfig() },
			override: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = testExpectedProtoc
				res.Proto.Dir = testExpectedProtoDir
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.Name = testExpectedName
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			isSet: func(string) bool { return true },
		},
		{
			desc: "not set",
			base: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = testExpectedProtoc
				res.DB.Driver = testExpectedDriver
				res.DB.Host = testExpectedHost
				res.DB.Port = testExpectedPort
				res.DB.UserName = testExpectedUserName
				res.DB.Password = testExpectedPassword
				return res
			},
			override: func() *rawConfig {
				res := newRawConfig()
				res.Proto.C = "foo"
				res.Proto.Dir = testExpectedProtoDir
				res.DB.Host = ""
				res.DB.Port = -1
				res.DB.Name = testExpectedName
				return res
			},
			isSet: func(name string) bool {
				switch name {
				case flagProtoDir, flagName:
					return true
				default:
					return false
				}
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			base := test.base()
			override := test.override()
			base.merge(override, test.isSet)
			if err := testRawConfigCheck(base); err != nil {
				t.Error(err)
			}
		})
	}
}
