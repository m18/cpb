package protos

import (
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/m18/cpb/check"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const testProtoc = "protoc"

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

func TestFileDescriptorSetBytes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "test", "proto") // relative to the test file location
	files := []string{
		filepath.Join(dir, "foo.proto"),
		filepath.Join(dir, "nested", "bar.proto"),
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
				protoc: testProtoc,
				dir:    test.dir,
			}
			res, err := p.fileDescriptorSetBytes(test.files)
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

func TestRegisterFileDescriptorSet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "test", "proto") // relative to the test file location
	files := []string{
		filepath.Join(dir, "foo_lite.proto"),
		filepath.Join(dir, "nested", "bar_lite.proto"),
	}
	p := &Protos{
		protoc: testProtoc,
		dir:    dir,
	}
	fdsb, err := p.fileDescriptorSetBytes(files)
	if err != nil {
		t.Fatalf("could not create file descriptor set: %v", err)
	}
	tests := []struct {
		desc  string
		fsdb  []byte
		files []string
		err   bool
	}{
		{
			desc: "nil bytes",
			fsdb: nil,
		},
		{
			desc: "empty bytes",
			fsdb: []byte{},
		},
		{
			desc: "garbage bytes",
			fsdb: []byte{1, 2, 3},
			err:  true,
		},
		{
			desc: "valid input",
			fsdb: fdsb,
			files: []string{
				"foo_lite.proto",
				filepath.Join("nested", "bar_lite.proto"),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p = &Protos{
				fileReg: &protoregistry.Files{},
			}
			err := p.registerFileDescriptorSet(test.fsdb)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if p.fileReg.NumFiles() != len(test.files) {
				t.Fatalf("expected number of files to be %d but it was %d", len(test.files), p.fileReg.NumFiles())
			}
			for _, file := range test.files {
				fd, err := p.fileReg.FindFileByPath(file)
				if err != nil {
					t.Fatal(err)
				}
				if fd == nil {
					t.Fatalf("expected file descriptor for %q to not be nil but it was", file)
				}
			}
		})
	}
}
