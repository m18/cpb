package protos

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const protoExt = ".proto"

// Protos performs protobuf-related operations
type Protos struct {
	protoc  string
	dir     string
	makeFS  func(string) fs.FS
	fileReg *protoregistry.Files
	mute    bool
}

// New returns a new Protos performing operations with protobuf types under dir
func New(protoc, dir string, makeFS func(string) fs.FS, mute bool) (*Protos, error) {
	res := &Protos{
		protoc:  protoc,
		dir:     dir,
		makeFS:  makeFS,
		fileReg: protoregistry.GlobalFiles,
		mute:    mute,
	}
	// if err := res.regFiles(); err != nil {
	// 	return nil, err
	// }
	return res, nil
}

func (p *Protos) registerFiles() error {
	if p.dir == "" {
		// dir was not provided -- no intent to query protobufs
		return nil
	}
	files, err := p.files()
	if err != nil {
		return fmt.Errorf("could not read dir %q: %w", p.dir, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("dir %q does not contain %s files", protoExt, p.dir)
	}
	fdsb, err := p.fileDescriptorSetBytes(files)
	if err != nil {
		return err
	}
	return p.registerFileDescriptorSet(fdsb)
}

func (p *Protos) files() ([]string, error) {
	res := []string{}
	fsys := p.makeFS(p.dir)
	// TODO: context
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
		[]string{"-I", p.dir, "--descriptor_set_out", os.Stdout.Name()},
		files...,
	)
	buf := &bytes.Buffer{}
	cmd := exec.Command(p.protoc, args...) // TODO: exec.CommandContext()
	cmd.Stdout = buf
	if p.mute {
		cmd.Stderr = io.Discard
	} else {
		cmd.Stderr = os.Stderr
	}
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
