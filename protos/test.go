package protos

import (
	"fmt"

	"github.com/m18/cpb/internal/testproto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func makeTestProtosLite() (*Protos, error) {
	return makeTestProtos(testproto.DirLite)
}

func makeTestProtos(dir string) (*Protos, error) {
	return New(
		testproto.Protoc,
		dir,
		testproto.Deterministic,
		testproto.MakeFS,
		testproto.MakeFileReg(),
		testproto.Mute,
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
