package testprotos

import (
	"github.com/m18/cpb/internal/testproto"
	"github.com/m18/cpb/protos"
)

func MakeProtosLite() (*protos.Protos, error) {
	return protos.New(
		testproto.Protoc,
		testproto.DirLite,
		testproto.Deterministic,
		testproto.MakeFS,
		testproto.MakeFileReg(),
		testproto.Mute,
	)
}
