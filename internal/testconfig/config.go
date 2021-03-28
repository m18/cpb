package testconfig

import (
	"fmt"
	"io/fs"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/testfs"
	"github.com/m18/cpb/internal/testproto"
)

const testConfigFormat = `{
	"protoc": "%s",
	"db": {
		"driver": "%s",
		"host": "localhost",
		"port": 1,
		"name": "cpb",
		"userName": "cpb",
		"password": "cpb",
		"params": {
			"foo": "bar"
		}
	},
	"messages": {
		"in": {
			"foo(id, text)": {
				"name": "testproto.lite.Foo",
				"template": {
					"id": "$id",
					"text": "$text"
				}
			},
			"bar(id, name)": {
				"name": "testproto.lite.nested.Bar",
				"template": {
					"id": "$id",
					"nested": {
						"name": "$name"
					}
				}
			}
		},
		"out": {
			"bar": {
				"name": "example.Bar"
			}
		}
	}
}`

func MakeTestConfigLite(driver string) (*config.Config, error) {
	jsn := fmt.Sprintf(testConfigFormat, testproto.Protoc, driver)
	testFS, testFileName := testfs.MakeTestConfigFS(jsn)
	return config.New([]string{"-" + config.FlagFile, testFileName}, func(string) fs.FS { return testFS })
}
