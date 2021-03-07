package internal

import (
	"io/fs"
	"testing/fstest"
)

const testConfigFileName = "my-config.json"

// MakeTestConfigFS creates an FS for use in unit tests
func MakeTestConfigFS(configFileContents string) (fs.FS, string) {
	file := &fstest.MapFile{
		Data: []byte(configFileContents),
	}
	dfs := fstest.MapFS{
		testConfigFileName: file,
	}
	return dfs, testConfigFileName
}
