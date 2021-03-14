package protos

import (
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/m18/cpb/check"
)

func TestProtosFiles(t *testing.T) {
	dir := &fstest.MapFile{Mode: fs.ModeDir}
	file := &fstest.MapFile{}
	tests := []struct {
		desc     string
		fsys     fs.FS
		dir      string
		expected []string
		err      bool
	}{
		{
			desc: "empty dir name",
			fsys: fstest.MapFS{},
			err:  true,
		},
		{
			desc: "empty dir",
			fsys: fstest.MapFS{"foo": dir},
			dir:  "foo",
		},
		{
			desc:     "single file",
			fsys:     fstest.MapFS{"foo/bar.proto": file},
			dir:      "foo",
			expected: []string{"foo/bar.proto"},
		},
		{
			desc:     "single file, nested dir",
			fsys:     fstest.MapFS{"foo/bar/baz.proto": file},
			dir:      "foo/bar",
			expected: []string{"foo/bar/baz.proto"},
		},
		{
			desc: "multiple files",
			fsys: fstest.MapFS{
				"foo/bar.proto": file,
				"foo/baz.proto": file,
			},
			dir: "foo",
			expected: []string{
				"foo/bar.proto",
				"foo/baz.proto",
			},
		},
		{
			desc: "multiple files, nested dir",
			fsys: fstest.MapFS{
				"foo/bar.proto":     file,
				"foo/baz/qux.proto": file,
			},
			dir: "foo",
			expected: []string{
				"foo/bar.proto",
				"foo/baz/qux.proto",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p := &Protos{
				dir:    test.dir,
				makeFS: func(string) fs.FS { return test.fsys },
			}
			res, err := p.files()
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if !check.StringSlicesAreEqual(res, test.expected) {
				t.Fatalf("expected %v but got %v", test.expected, res)
			}
		})
	}
}

func TestDescriptorSetBytes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "example", "proto") // relative to the test file location
	files := []string{
		filepath.Join(dir, "address_book.proto"),
		filepath.Join(dir, "nested", "note.proto"),
	}
	tests := []struct {
		desc  string
		dir   string
		files []string
		err   bool
	}{
		{
			desc:  "nil files",
			dir:   dir,
			files: nil,
			err:   true,
		},
		{
			desc:  "empty files",
			dir:   dir,
			files: []string{},
			err:   true,
		},
		{
			desc:  "valid input",
			dir:   dir,
			files: files,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p := &Protos{
				protoc: "protoc",
				dir:    test.dir,
			}
			res, err := p.descriptorSetBytes(test.files)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if res == nil || len(res) == 0 {
				t.Fatalf("expected result not to be empty but it was")
			}
		})
	}
}
