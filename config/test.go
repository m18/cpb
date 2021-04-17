package config

import "fmt"

const (
	testExpectedProtoc   = "prc"
	testExpectedProtoDir = "pr/d"
	testExpectedDriver   = "drv"
	testExpectedHost     = "hst"
	testExpectedPort     = 5500
	testExpectedName     = "db"
	testExpectedUserName = "unm"
	testExpectedPassword = "pwd"

	testQuery = "select * from foo;"
)

var testConfigJSON = fmt.Sprintf(`{
	"proto": {
		"c": "%s",
		"dir": "%s"
	},
	"db": {
		"driver": "%s",
		"host": "%s",
		"port": %d,
		"name": "%s",
		"userName": "%s",
		"password": "%s",
		"params": {
			"foo": "bar"
		}
	},
	"messages": {
		"in": {
			"foo()": {
				"name": "example.Foo"
			}
		},
		"out": {
			"bar": {
				"name": "example.Bar"
			}
		}
	}
}`, testExpectedProtoc, testExpectedProtoDir, testExpectedDriver, testExpectedHost, testExpectedPort, testExpectedName, testExpectedUserName, testExpectedPassword)

func testRawConfigCheck(c *rawConfig) error {
	if c.Proto.C != testExpectedProtoc {
		return fmt.Errorf("expected protoc to be %q but it was %q", testExpectedProtoc, c.Proto.C)
	}
	if c.Proto.Dir != testExpectedProtoDir {
		return fmt.Errorf("expected proto dir to be %q but it was %q", testExpectedProtoDir, c.Proto.Dir)
	}
	if c.DB.Driver != testExpectedDriver {
		return fmt.Errorf("expected driver to be %q but it was %q", testExpectedDriver, c.DB.Driver)
	}
	if c.DB.Host != testExpectedHost {
		return fmt.Errorf("expected host to be %q but it was %q", testExpectedHost, c.DB.Host)
	}
	if c.DB.Port != testExpectedPort {
		return fmt.Errorf("expected port to be %d but it was %d", testExpectedPort, c.DB.Port)
	}
	if c.DB.Name != testExpectedName {
		return fmt.Errorf("expected name to be %s but it was %s", testExpectedName, c.DB.Name)
	}
	if c.DB.UserName != testExpectedUserName {
		return fmt.Errorf("expected user name to be %q but it was %q", testExpectedUserName, c.DB.UserName)
	}
	if c.DB.Password != testExpectedPassword {
		return fmt.Errorf("expected password to be %q but it was %q", testExpectedPassword, c.DB.Password)
	}
	return nil
}

func testConfigCheck(c *Config) error {
	if c.Proto.C != testExpectedProtoc {
		return fmt.Errorf("expected protoc to be %q but it was %q", testExpectedProtoc, c.Proto.C)
	}
	if c.DB.Driver != testExpectedDriver {
		return fmt.Errorf("expected driver to be %q but it was %q", testExpectedDriver, c.DB.Driver)
	}
	if c.DB.Host != testExpectedHost {
		return fmt.Errorf("expected host to be %q but it was %q", testExpectedHost, c.DB.Host)
	}
	if c.DB.Port != testExpectedPort {
		return fmt.Errorf("expected port to be %d but it was %d", testExpectedPort, c.DB.Port)
	}
	if c.DB.Name != testExpectedName {
		return fmt.Errorf("expected name to be %s but it was %s", testExpectedName, c.DB.Name)
	}
	if c.DB.UserName != testExpectedUserName {
		return fmt.Errorf("expected user name to be %q but it was %q", testExpectedUserName, c.DB.UserName)
	}
	if c.DB.Password != testExpectedPassword {
		return fmt.Errorf("expected password to be %q but it was %q", testExpectedPassword, c.DB.Password)
	}
	return nil
}
