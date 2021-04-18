package protos

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"text/template"

	"github.com/m18/cpb/check"
	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/testproto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

func TestProtosNew(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "testproto")
	tests := []struct {
		desc string
		dir  string
		err  bool
	}{
		{
			desc: "empty dir name",
			dir:  "",
		},
		{
			desc: "empty dir",
			dir:  filepath.Join(dir, "empty"),
			err:  true,
		},
		{
			desc: "valid dir",
			dir:  filepath.Join(dir, "lite"),
		},
		{
			desc: "invalid proto",
			dir:  filepath.Join(dir, "invalid"),
			err:  true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p, err := makeTestProtos(test.dir)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if p == nil {
				t.Fatalf("expected New to not return nil but it did")
			}
		})
	}
}

func TestProtosProtoBytes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "testproto")
	tests := []struct {
		desc     string
		dir      string
		message  protoreflect.FullName
		fromJSON string
		err      bool
	}{
		{
			desc:    "empty registry",
			dir:     "",
			message: "testproto.lite.Foo",
			err:     true,
		},
		{
			desc:     "valid input",
			dir:      filepath.Join(dir, "lite"),
			message:  "testproto.lite.Foo",
			fromJSON: `{"id": 5, "text": "five"}`,
		},
		{
			desc:    "empty message",
			dir:     filepath.Join(dir, "lite"),
			message: "",
			err:     true,
		},
		{
			desc:     "invalid fromJSON",
			dir:      filepath.Join(dir, "lite"),
			message:  "testproto.lite.Foo",
			fromJSON: `{"id": 5, "invalid": "five"}`,
			err:      true,
		},
		{
			desc:    "unknown message",
			dir:     filepath.Join(dir, "lite"),
			message: "testproto.lite.Qux",
			err:     true,
		},
		{
			desc:    "non-message",
			dir:     filepath.Join(dir, "lite"),
			message: "testproto.lite.nested.Bar.Qux",
			err:     true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p, err := makeTestProtos(test.dir)
			if err != nil {
				t.Fatalf("expected makeTestProtos to not return error but it did")
			}
			b, err := p.ProtoBytes(test.message, test.fromJSON)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if b == nil || len(b) == 0 {
				t.Fatalf("expected ProtoBytes result to not be empty but it was")
			}
		})
	}
}

func TestProtosStringerFor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	const missingTplValue = "<no value>"
	validom := &config.OutMessage{
		Name:     "testproto.lite.nested.Bar",
		Template: template.Must(template.New("").Parse(`Hello, {{.text}}! Hi, {{.nested_name}}.`)),
		Props:    map[string]struct{}{"text": {}, "nested.name": {}},
	}
	p, err := makeTestProtosLite()
	if err != nil {
		t.Fatalf("expected makeTestProtosLite to not return error but it did")
	}
	validb, err := p.ProtoBytes(validom.Name, `{"id": 5, "text": "world", "nested": {"name": "cosmos"}}`)
	if err != nil {
		t.Fatalf("expected ProtoBytes to not return error but it did")
	}
	tests := []struct {
		desc        string
		om          *config.OutMessage
		b           []byte
		expected    string
		err         bool
		stringerErr bool
	}{
		{
			desc:     "valid input",
			om:       validom,
			b:        validb,
			expected: "Hello, world! Hi, cosmos.",
		},
		{
			desc: "fewer props (not possible due to config parsing)",
			om: &config.OutMessage{
				Name:     validom.Name,
				Template: validom.Template,
				Props:    map[string]struct{}{"text": {}},
			},
			b:        validb,
			expected: fmt.Sprintf("Hello, world! Hi, %s.", missingTplValue),
		},
		{
			desc: "more props (not possible due to config parsing)",
			om: &config.OutMessage{
				Name:     validom.Name,
				Template: validom.Template,
				Props:    map[string]struct{}{"id": {}, "text": {}, "nested.name": {}},
			},
			b:        validb,
			expected: "Hello, world! Hi, cosmos.",
		},
		{
			desc: "non-existent message name",
			om:   &config.OutMessage{Name: "nonexistent"},
			err:  true,
		},
		{
			// TODO: support optional props (e.g., versioning -- a prop exists only in a newer version of a message)
			desc: "non-existent prop name",
			om: &config.OutMessage{
				Name:  validom.Name,
				Props: map[string]struct{}{"nonexistent": {}},
			},
			err: true,
		},
		{
			desc:        "invalid bytes",
			om:          validom,
			b:           []byte{1, 2, 3},
			stringerErr: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			stringer, err := p.StringerFor(test.om)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			str, err := stringer(test.b)
			if err == nil == test.stringerErr {
				t.Fatalf("expected %t but did not get it: %v", test.stringerErr, err)
			}
			if test.stringerErr {
				return
			}
			if str != test.expected {
				t.Fatalf("expected %q but got %q", test.expected, str)
			}
		})
	}
}

func TestProtosMessageDescriptor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	p, err := makeTestProtosLite()
	if err != nil {
		t.Fatalf("expected makeTestProtosLite to not return error but it did")
	}
	validFileReg := p.fileReg
	tests := []struct {
		desc    string
		fileReg *protoregistry.Files
		message protoreflect.FullName
		err     bool
	}{
		{
			desc:    "empty registry",
			fileReg: &protoregistry.Files{},
			message: "testproto.lite.Foo",
			err:     true,
		},
		{
			desc:    "empty message name",
			fileReg: validFileReg,
			message: "",
			err:     true,
		},
		{
			desc:    "non-message",
			fileReg: validFileReg,
			message: "testproto.lite.nested.Bar.Qux",
			err:     true,
		},
		{
			desc:    "valid input",
			fileReg: validFileReg,
			message: "testproto.lite.Foo",
		},
		{
			desc:    "nested dir",
			fileReg: validFileReg,
			message: "testproto.lite.nested.Bar",
		},
		{
			desc:    "nested message",
			fileReg: validFileReg,
			message: "testproto.lite.nested.Bar.Baz",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p := &Protos{fileReg: test.fileReg}
			md, err := p.messageDescriptor(test.message)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if md == nil {
				t.Fatalf("expected MessageDescriptor to not be empty but it was")
			}
		})
	}
}

func TestProtosRegisterFiles(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "testproto")
	makeFS := func(dir string) fs.FS { return os.DirFS(dir) }
	tests := []struct {
		desc          string
		dir           string
		expectedFiles []string
		err           bool
	}{
		{
			desc: "empty dir name",
			dir:  "",
		},
		{
			desc: "empty dir",
			dir:  filepath.Join(dir, "empty"),
			err:  true,
		},
		{
			desc: "non-existent dir",
			dir:  filepath.Join(dir, "nonexistent"),
			err:  true,
		},
		{
			desc: "invalid file",
			dir:  filepath.Join(dir, "invalid"),
			err:  true,
		},
		{
			desc: "valid dir",
			dir:  filepath.Join(dir, "lite"),
			expectedFiles: []string{
				"foo_lite.proto",
				"nested/bar_lite.proto",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			p := &Protos{
				protoc:  testproto.Protoc,
				dir:     test.dir,
				makeFS:  makeFS,
				fileReg: &protoregistry.Files{},
				mute:    true,
			}
			err := p.registerFiles()
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			for _, file := range test.expectedFiles {
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

func TestProtosFiles(t *testing.T) {
	dir := &fstest.MapFile{Mode: fs.ModeDir}
	file := &fstest.MapFile{}
	tests := []struct {
		desc     string
		fsys     fs.FS
		expected []string
		err      bool
	}{
		{
			desc: "empty fs",
			fsys: fstest.MapFS{},
		},
		{
			desc: "empty dir",
			fsys: fstest.MapFS{"foo": dir},
		},
		{
			desc:     "single file",
			fsys:     fstest.MapFS{"foo/bar.proto": file},
			expected: []string{"foo/bar.proto"},
		},
		{
			desc:     "single file, nested dir",
			fsys:     fstest.MapFS{"foo/bar/baz.proto": file},
			expected: []string{"foo/bar/baz.proto"},
		},
		{
			desc: "multiple files",
			fsys: fstest.MapFS{
				"foo/bar.proto": file,
				"foo/baz.proto": file,
			},
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

func TestProtosFileDescriptorSetBytes(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "testproto")
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
				protoc: testproto.Protoc,
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

func TestProtosRegisterFileDescriptorSet(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dir := filepath.Join("..", "internal", "testproto", "lite")
	files := []string{
		filepath.Join(dir, "foo_lite.proto"),
		filepath.Join(dir, "nested", "bar_lite.proto"),
	}
	p := &Protos{
		protoc: testproto.Protoc,
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
			p := &Protos{
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
