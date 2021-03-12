package internal

import (
	"io/fs"
	"testing/fstest"
)

const testConfigFileName = "my-config.json"

// MakeTestConfigFS creates an FS for use in unit tests
func MakeTestConfigFS(configFileContents string) (fs.FS, string) {
	return makeTestConfigFS(configFileContents, testConfigFileName)
}

// MakeTestConfigFSCustom creates an FS with a custom file name for use in unit tests
func MakeTestConfigFSCustom(configFileContents, fileName string) (fs.FS, string) {
	return makeTestConfigFS(configFileContents, fileName)
}

func makeTestConfigFS(configFileContents, fileName string) (fs.FS, string) {
	file := &fstest.MapFile{
		Data: []byte(configFileContents),
	}
	dfs := fstest.MapFS{
		fileName: file,
	}
	return dfs, fileName
}
