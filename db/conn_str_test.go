package db

import (
	"testing"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/testcheck"
)

func TestConnStrGens(t *testing.T) {
	validConfig := config.DBConfig{
		Driver:   DriverPostgres,
		Host:     "host.com",
		Port:     5555,
		Name:     "name",
		UserName: "userName",
		Password: "password",
		Params:   map[string]string{"key": "value"},
	}
	tests := []struct {
		desc            string
		makeConfig      func(config.DBConfig) *config.DBConfig
		expectedConnStr string
		err             bool
	}{
		{
			desc: "valid input",
			makeConfig: func(valid config.DBConfig) *config.DBConfig {
				return &valid
			},
			expectedConnStr: "postgres://userName:password@host.com:5555/name?key=value",
		},
		{
			desc: "valid, needs escaping",
			makeConfig: func(valid config.DBConfig) *config.DBConfig {
				valid.Name = "name!"
				valid.Password = "pass!word"
				valid.Params = map[string]string{"key": "va!ue"}
				return &valid
			},
			expectedConnStr: "postgres://userName:pass%21word@host.com:5555/name%21?key=va%21ue",
		},
		{
			desc: "valid, default port",
			makeConfig: func(valid config.DBConfig) *config.DBConfig {
				valid.Port = 0
				return &valid
			},
			expectedConnStr: "postgres://userName:password@host.com:5432/name?key=value",
		},
		{
			desc: "invalid URL (host)",
			makeConfig: func(valid config.DBConfig) *config.DBConfig {
				valid.Host = "!host.com"
				return &valid
			},
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			cfg := test.makeConfig(validConfig)
			gen := connStrGens[cfg.Driver]
			connStr, err := gen(cfg)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if connStr != test.expectedConnStr {
				t.Fatalf("expected %q but got %q", test.expectedConnStr, connStr)
			}
		})
	}
}
