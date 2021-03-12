package config

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"testing"

	"github.com/m18/cpb/config/internal"
)

func TestParserParseFlags(t *testing.T) {
	tests := []struct {
		args             []string
		expectedFileName string
		check            func(*rawConfig) error
		err              bool
	}{
		{
			expectedFileName: "",
		},
		{
			args:             []string{},
			expectedFileName: "",
		},
		{
			args:             []string{"-" + flagFile, ""},
			expectedFileName: "",
		},
		{
			args: []string{
				"-" + flagFile, defaultConfigFileName,
				"-" + flagProtoc, testExpectedProtoc,
				"-" + flagDriver, testExpectedDriver,
				"-" + flagHost, testExpectedHost,
				"-" + flagPort, strconv.Itoa(testExpectedPort),
				"-" + flagName, testExpectedName,
				"-" + flagUserName, testExpectedUserName,
				"-" + flaagPassword, testExpectedPassword,
			},
			expectedFileName: defaultConfigFileName,
			check:            testRawConfigCheck,
		},
		{
			args: []string{"-unknown"},
			err:  true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprint(test.args), func(t *testing.T) {
			t.Parallel()
			fileName, config, err := newParser(test.args, nil, true).parseFlags()
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if fileName != test.expectedFileName {
				t.Fatalf("expected file name to be %q but it was %q", test.expectedFileName, fileName)
			}
			if test.check == nil {
				return
			}
			if err := test.check(config); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestParserParseFile(t *testing.T) {
	testFS, testFileName := internal.MakeTestConfigFS(testConfigJSON)
	testMakeFS := func(string) fs.FS {
		return testFS
	}
	testFSDefault, _ := internal.MakeTestConfigFSCustom(testConfigJSON, defaultConfigFileName)
	testMakeFSDefault := func(string) fs.FS {
		return testFSDefault
	}
	tests := []struct {
		desc     string
		makeFS   func(string) fs.FS
		fileName string
		err      bool
		check    func(*rawConfig) error
	}{
		{
			desc:     "valid input",
			makeFS:   testMakeFS,
			fileName: testFileName,
			check:    testRawConfigCheck,
		},
		{
			desc:     "valid input, optional default config file",
			makeFS:   testMakeFSDefault,
			fileName: "",
			check:    testRawConfigCheck,
		},
		{
			desc:     "non-existent file",
			makeFS:   testMakeFS,
			fileName: "none.config",
			err:      true,
		},
		{
			desc:     "empty file name",
			makeFS:   testMakeFS,
			fileName: "",
			check: func(c *rawConfig) error {
				if c != nil {
					return fmt.Errorf("expected nil but did not get it")
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			raw, err := newParser(nil, test.makeFS, false).parseFile(test.fileName)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.check == nil {
				return
			}
			if err = test.check(raw); err != nil {
				t.Fatal(err)
			}
			if raw == nil {
				return
			}
			if raw.Messages.In == nil || len(raw.Messages.In) == 0 {
				t.Fatal("expected in message config to not be nil or empty but it was")
			}
			if raw.Messages.Out == nil || len(raw.Messages.Out) == 0 {
				t.Fatal("expected out message config to not be nil or empty but it was")
			}
		})
	}
}

func TestParserFrom(t *testing.T) {
	raw, err := newRawConfig().from(strings.NewReader(testConfigJSON))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := newParser(nil, nil, true).from(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err = testConfigCheck(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestParserParse(t *testing.T) {
	testFS, testFileName := internal.MakeTestConfigFS(testConfigJSON)
	_ = testFileName
	testMakeFS := func(string) fs.FS { return testFS }
	tests := []struct {
		args  []string
		check func(*Config) error
		err   bool
	}{
		// {
		// 	args: []string{"-unknown"},
		// 	err:  true,
		// },
		// {
		// 	args: []string{"-" + flagPort, "foo"},
		// 	err:  true,
		// },
		// {
		// 	args: []string{"-" + flagFile, "unknown.config"},
		// 	err:  true,
		// },
		{
			args:  []string{"-" + flagFile, testFileName},
			check: testConfigCheck,
		},
		// {
		// 	args: []string{
		// 		"-" + flagFile, testFileName,
		// 		"-" + flagProtoc, "foo",
		// 		"-" + flagDriver, "bar",
		// 	},
		// 	check: func(c *Config) error {
		// 		if c.Protoc != "foo" {
		// 			return fmt.Errorf("expected protoc to be %q but it was %q", "foo", c.Protoc)
		// 		}
		// 		if c.DB.Driver != "bar" {
		// 			return fmt.Errorf("expected driver to be %q but it was %q", "bar", c.DB.Driver)
		// 		}
		// 		return nil
		// 	},
		// },
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			t.Parallel()
			cfg, err := newParser(test.args, testMakeFS, true).parse()
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err || test.check == nil {
				return
			}
			if err = test.check(cfg); err != nil {
				t.Fatal(err)
			}
			if cfg.InMessages == nil || len(cfg.InMessages) == 0 {
				t.Fatal("expected in message config to not be nil or empty but it was")
			}
			if cfg.OutMessages == nil || len(cfg.OutMessages) == 0 {
				t.Fatal("expected out message config to not be nil or empty but it was")
			}
		})
	}
}
