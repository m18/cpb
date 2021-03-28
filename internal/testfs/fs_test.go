package testfs

import (
	"io"
	"io/fs"
	"testing"
)

func TestMakeTestConfigFS(t *testing.T) {
	contents := "foo"
	tests := []struct {
		useDefautTestFileName bool
		customFileName        string
	}{
		{useDefautTestFileName: true},
		{customFileName: "bar.json"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.customFileName, func(t *testing.T) {
			t.Parallel()
			var fsys fs.FS
			var fileName string
			var expectedFileName string
			if test.useDefautTestFileName {
				fsys, fileName = MakeTestConfigFS(contents)
				expectedFileName = testConfigFileName
			} else {
				fsys, fileName = MakeTestConfigFSCustom(contents, test.customFileName)
				expectedFileName = test.customFileName
			}
			if fileName != expectedFileName {
				t.Fatalf("expected file name to be %q but it was %q", expectedFileName, fileName)
			}
			f, err := fsys.Open(fileName)
			if err != nil {
				t.Fatal(err)
			}
			if f == nil {
				t.Fatalf("expected a file but got nil")
			}
			bread, err := io.ReadAll(f)
			if err != nil {
				t.Fatal(err)
			}
			read := string(bread)
			if read != contents {
				t.Fatalf("expected %q but got %q", contents, read)
			}
		})
	}
}
