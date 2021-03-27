package testproto

import (
	"io/fs"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	Protoc = "protoc"
	Mute   = true
)

var (
	DirLite     = filepath.Join("..", "internal", "testproto", "lite")
	MakeFS      = func(dir string) fs.FS { return os.DirFS(dir) }
	MakeFileReg = func() *protoregistry.Files { return &protoregistry.Files{} }
)
