package protos

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	descriptorSetOut = "/dev/stdout" // TODO: x-platform
	protoExt         = ".proto"
)

// Protos performs protobuf-related operations
type Protos struct {
	protoc  string
	dir     string
	makeFS  func(string) fs.FS
	fileReg *protoregistry.Files
}

// New returns a new Protos performing operations with protobuf types under dir
func New(protoc, dir string, makeFS func(string) fs.FS) (*Protos, error) {
	res := &Protos{
		protoc:  protoc,
		dir:     dir,
		makeFS:  makeFS,
		fileReg: protoregistry.GlobalFiles,
	}
	// if err := res.regFiles(); err != nil {
	// 	return nil, err
	// }
	return res, nil
}

// TODO: it's OK for "dir" to be empty; in this case ignore protobuf completely
//       but if it's not empty, then any issue is an error (e.g., dir has no files is an error)

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

func (p *Protos) fileDescriptorSetBytes(files []string) ([]byte, error) {
	args := append(
		[]string{"-I", p.dir, "--descriptor_set_out", descriptorSetOut},
		files...,
	)
	buf := &bytes.Buffer{}
	cmd := exec.Command(p.protoc, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Protos) registerFileDescriptorSet(fdsb []byte) error {
	fds := &descriptorpb.FileDescriptorSet{}
	if err := proto.Unmarshal(fdsb, fds); err != nil {
		return err
	}
	for _, fdp := range fds.GetFile() {
		fd, err := protodesc.NewFile(fdp, p.fileReg)
		if err != nil {
			return err
		}
		if err = p.fileReg.RegisterFile(fd); err != nil {
			return err
		}
	}
	return nil
}
