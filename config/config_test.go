package config

import (
	"fmt"
	"io/fs"
	"testing"
	"text/template"

	"github.com/m18/cpb/internal/testcheck"
	"github.com/m18/cpb/internal/testfs"
)

func TestConfigNew(t *testing.T) {
	testFS, testFileName := testfs.MakeTestConfigFS(testConfigJSON)
	testMakeFS := func(string) fs.FS { return testFS }
	testFSDefault, _ := testfs.MakeTestConfigFSCustom(testConfigJSON, defaultConfigFileName)
	testMakeFSDefault := func(string) fs.FS { return testFSDefault }
	tests := []struct {
		args   []string
		makeFS func(string) fs.FS
		err    bool
	}{
		{
			args:   []string{"-" + FlagFile, "unknown.config"},
			makeFS: testMakeFS,
			err:    true,
		},
		{
			args: []string{
				"-" + flagPort, "0",
			},
			makeFS: testMakeFS,
			err:    true,
		},
		{
			args:   []string{"-" + FlagFile, testFileName},
			makeFS: testMakeFS,
		},
		{
			args:   nil,
			makeFS: testMakeFSDefault,
		},
		{
			args:   []string{},
			makeFS: testMakeFSDefault,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			t.Parallel()
			cfg, err := New(test.args, test.makeFS)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if !test.err && cfg == nil {
				t.Fatalf("expected config to not be nil but it was")
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	makeCfg := func(upd func(*Config)) *Config {
		res := &Config{
			Proto: &Proto{
				C: testExpectedProtoc,
			},
			DB: &DBConfig{
				Driver:   testExpectedDriver,
				Host:     testExpectedHost,
				Port:     testExpectedPort,
				Name:     testExpectedName,
				UserName: testExpectedUserName,
			},
		}
		if upd != nil {
			upd(res)
		}
		return res
	}
	tests := []struct {
		desc string
		upd  func(*Config)
		err  bool
	}{
		{
			desc: "valid",
		},
		{
			desc: "no protoc, default is used",
			upd:  func(c *Config) { c.Proto.C = "" },
		},
		{
			desc: "no driver",
			upd:  func(c *Config) { c.DB.Driver = "" },
			err:  true,
		},
		{
			desc: "no host",
			upd:  func(c *Config) { c.DB.Host = "" },
			err:  true,
		},
		{
			desc: "no port",
			upd:  func(c *Config) { c.DB.Port = 0 },
			err:  true,
		},
		{
			desc: "no database name",
			upd:  func(c *Config) { c.DB.Name = "" },
			err:  true,
		},
		{
			desc: "no user name",
			upd:  func(c *Config) { c.DB.UserName = "" },
			err:  true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			testcheck.FatalIfUnexpected(t, makeCfg(test.upd).validate(), test.err)
		})
	}
}

func TestInMessageJSON(t *testing.T) {
	tpl0 := "hello, world! hi, cosmos."
	tpl2 := "hello, {{.foo}}! hi, {{.bar}}."
	tests := []struct {
		desc     string
		im       *InMessage
		args     []string
		expected string
		err      bool
	}{
		{
			desc: "no params, no args",
			im: &InMessage{
				template: template.Must(template.New("").Parse(tpl0)),
			},
			expected: tpl0,
		},
		{
			desc: "2 params, 2 args",
			im: &InMessage{
				template: template.Must(template.New("").Parse(tpl2)),
				params:   []string{"foo", "bar"},
			},
			args:     []string{"world", "cosmos"},
			expected: tpl0,
		},
		{
			desc: "2 params, 2 args, different order",
			im: &InMessage{
				template: template.Must(template.New("").Parse(tpl2)),
				params:   []string{"foo", "bar"},
			},
			args:     []string{"cosmos", "world"},
			expected: "hello, cosmos! hi, world.",
		},
		{
			desc: "0 params, 2 args",
			im: &InMessage{
				template: template.Must(template.New("").Parse(tpl0)),
			},
			args: []string{"world", "cosmos"},
			err:  true,
		},
		{
			desc: "2 params, 0 args",
			im: &InMessage{
				template: template.Must(template.New("").Parse(tpl2)),
				params:   []string{"foo", "bar"},
			},
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			res, err := test.im.JSON(test.args)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if res != test.expected {
				t.Fatalf("expected %q but got %q", test.expected, res)
			}
		})
	}
}
