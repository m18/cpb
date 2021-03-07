package internal

import (
	"io"
	"testing"
)

func TestMakeTestConfigFS(t *testing.T) {
	contents := "foo"
	dfs, fileName := MakeTestConfigFS(contents)
	if fileName != testConfigFileName {
		t.Fatalf("expected file name to be %q but it was %q", testConfigFileName, fileName)
	}
	f, err := dfs.Open(fileName)
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
}
