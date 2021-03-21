package protos

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

const testProtoc = "protoc"

func makeTestProtosLite() (*Protos, error) {
	dir := filepath.Join("..", "internal", "test", "proto", "lite")
	return makeTestProtos(dir)
}

func makeTestProtos(dir string) (*Protos, error) {
	return New(
		testProtoc,
		dir,
		func(dir string) fs.FS { return os.DirFS(dir) },
		&protoregistry.Files{},
		true,
	)
}

func barLiteMessageDescriptor() (protoreflect.MessageDescriptor, error) {
	p, err := makeTestProtosLite()
	if err != nil {
		return nil, fmt.Errorf("error calling makeTestProtosLite: %w", err)
	}
	const name = "testproto.lite.nested.Bar"
	d, err := p.fileReg.FindDescriptorByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not find message %q: %w", name, err)
	}
	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a MessageDescritor: %q", name)
	}
	return md, nil
}
