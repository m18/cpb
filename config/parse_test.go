package config

import (
	"fmt"
	"io/fs"
	"strconv"
	"testing"

	"github.com/m18/cpb/internal/testcheck"
	"github.com/m18/cpb/internal/testfs"
)

func TestParserParseCLArgs(t *testing.T) {
	tests := []struct {
		args             []string
		expectedFilePath string
		expectedQuery    string
		check            func(*rawConfig) error
		err              bool
	}{
		{
			expectedFilePath: "",
		},
		{
			args:             []string{},
			expectedFilePath: "",
		},
		{
			args:             []string{"-" + FlagFile, ""},
			expectedFilePath: "",
		},
		{
			args: []string{
				"-" + FlagFile, defaultConfigFileName,
				"-" + flagProtoc, testExpectedProtoc,
				"-" + flagProtoDir, testExpectedProtoDir,
				"-" + flagDriver, testExpectedDriver,
				"-" + flagHost, testExpectedHost,
				"-" + flagPort, strconv.Itoa(testExpectedPort),
				"-" + flagName, testExpectedName,
				"-" + flagUserName, testExpectedUserName,
				"-" + flagPassword, testExpectedPassword,
			},
			expectedFilePath: defaultConfigFileName,
			check:            testRawConfigCheck,
		},
		{
			args:          []string{testQuery},
			expectedQuery: testQuery,
		},
		{
			args: []string{
				"-" + FlagFile, defaultConfigFileName,
				testQuery,
			},
			expectedFilePath: defaultConfigFileName,
			expectedQuery:    testQuery,
		},
		{
			args: []string{
				testQuery,
				"-" + FlagFile, defaultConfigFileName,
			},
			expectedFilePath: "",
			expectedQuery:    testQuery,
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
			filePath, config, flagSet, err := newParser(test.args, nil, true).parseCLArgs()
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if filePath != test.expectedFilePath {
				t.Fatalf("expected file name to be %q but it was %q", test.expectedFilePath, filePath)
			}
			if config.DB.Query != test.expectedQuery {
				t.Fatalf("expected query to be %q but it was %q", test.expectedQuery, config.DB.Query)
			}
			if flagSet == nil {
				t.Fatalf("expected flagSet to not be nil but it was")
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
	testFS, testFileName := testfs.MakeTestConfigFS(testConfigJSON)
	testMakeFS := func(string) fs.FS {
		return testFS
	}
	testFSDefault, _ := testfs.MakeTestConfigFSCustom(testConfigJSON, defaultConfigFileName)
	testMakeFSDefault := func(string) fs.FS {
		return testFSDefault
	}
	tests := []struct {
		desc     string
		makeFS   func(string) fs.FS
		filePath string
		isSet    bool
		err      bool
		check    func(*rawConfig) error
	}{
		{
			desc:     "valid input",
			makeFS:   testMakeFS,
			filePath: testFileName,
			isSet:    true,
			check:    testRawConfigCheck,
		},
		{
			desc:     "valid input, optional default config file",
			makeFS:   testMakeFSDefault,
			filePath: "foo",
			isSet:    false,
			check:    testRawConfigCheck,
		},
		{
			desc:     "non-existent file",
			makeFS:   testMakeFS,
			filePath: "none.config",
			isSet:    true,
			err:      true,
		},
		{
			desc:   "no config file at all",
			makeFS: testMakeFS,
			isSet:  false,
			check: func(c *rawConfig) error {
				if c == nil {
					return fmt.Errorf("expected config to not be nil but it was")
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			raw, err := newParser(nil, test.makeFS, false).parseFile(test.filePath, test.isSet)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			testcheck.FatalIf(t, test.check(raw))
			if !test.isSet {
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
	raw := newRawConfig()
	err := raw.from([]byte(testConfigJSON))
	testcheck.FatalIf(t, err)
	cfg, err := newParser(nil, nil, true).from(raw)
	testcheck.FatalIf(t, err)
	testcheck.FatalIf(t, testConfigCheck(cfg))
}

func TestParserParse(t *testing.T) {
	testFS, testFileName := testfs.MakeTestConfigFS(testConfigJSON)
	_ = testFileName
	testMakeFS := func(string) fs.FS { return testFS }
	tests := []struct {
		args  []string
		check func(*Config) error
		err   bool
	}{
		{
			args: []string{"-unknown"},
			err:  true,
		},
		{
			args: []string{"-" + flagPort, "foo"},
			err:  true,
		},
		{
			args: []string{"-" + FlagFile, "unknown.config"},
			err:  true,
		},
		{
			args:  []string{"-" + FlagFile, testFileName},
			check: testConfigCheck,
		},
		{
			args: []string{
				"-" + FlagFile, testFileName,
				"-" + flagProtoc, "foo",
				"-" + flagProtoDir, "bar",
				"-" + flagDriver, "baz",
				"-" + flagNoAutoMap,
				"-" + flagUndeterministic,
			},
			check: func(c *Config) error {
				if c.Proto.C != "foo" {
					return fmt.Errorf("expected protoc to be %q but it was %q", "foo", c.Proto.C)
				}
				if c.Proto.Dir != "bar" {
					return fmt.Errorf("expected proto dir to be %q but it was %q", "bar", c.Proto.Dir)
				}
				if c.DB.Driver != "baz" {
					return fmt.Errorf("expected driver to be %q but it was %q", "baz", c.DB.Driver)
				}
				if c.AutoMapOutMessages {
					return fmt.Errorf("expected auto-map to be false but it was not")
				}
				if c.Proto.Deterministic {
					return fmt.Errorf("expected deterministic to be false but it was not")
				}
				return nil
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("%v", test.args), func(t *testing.T) {
			t.Parallel()
			cfg, err := newParser(test.args, testMakeFS, true).parse()
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err || test.check == nil {
				return
			}
			testcheck.FatalIf(t, test.check(cfg))
			if cfg.InMessages == nil || len(cfg.InMessages) == 0 {
				t.Fatal("expected in message config to not be nil or empty but it was")
			}
			if cfg.OutMessages == nil || len(cfg.OutMessages) == 0 {
				t.Fatal("expected out message config to not be nil or empty but it was")
			}
		})
	}
}
