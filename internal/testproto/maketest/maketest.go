package maketest

import (
	"github.com/m18/cpb/internal/testproto"
	"github.com/m18/cpb/protos"
)

func ProtosLite() (*protos.Protos, error) {
	return protos.New(
		testproto.Protoc,
		testproto.DirLite,
		testproto.MakeFS,
		testproto.MakeFileReg(),
		testproto.Mute,
	)
}
