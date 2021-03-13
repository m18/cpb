package protos

import (
	"io/fs"
	"strings"

	"google.golang.org/protobuf/reflect/protoregistry"
)

const (
	dsTempFileName = ".tmp_proto.ds"
	protoExt       = ".proto"
)

// Protos performs protobuf-related operations
type Protos struct {
	protoc         string
	dir            string
	makeFS         func(string) fs.FS
	dsTempFileName string
	fileReg        *protoregistry.Files
}

// New returns a new Protos performing operations with protobuf types under dir
func New(protoc, dir string, makeFS func(string) fs.FS) (*Protos, error) {
	res := &Protos{
		protoc:         protoc,
		dir:            dir,
		makeFS:         makeFS,
		dsTempFileName: dsTempFileName,
		fileReg:        protoregistry.GlobalFiles,
	}
	// if err := res.regFiles(); err != nil {
	// 	return nil, err
	// }
	return res, nil
}

func (p *Protos) files() ([]string, error) {
	res := []string{}
	fsys := p.makeFS(p.dir)
	err := fs.WalkDir(fsys, p.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == "" {
			return nil
		}
		if strings.HasSuffix(d.Name(), protoExt) {
			res = append(res, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
