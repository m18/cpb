package protos

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"github.com/m18/cpb/config"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

const protoExt = ".proto"

// Protos performs protobuf-related operations.
type Protos struct {
	protoc  string
	dir     string
	makeFS  func(string) fs.FS
	fileReg *protoregistry.Files
	mute    bool
}

// New returns a new Protos performing operations with protobuf types under dir.
//
// dir can be an empty string, which implies that there is no intent to query protobufs.
func New(protoc, dir string, makeFS func(string) fs.FS, fileReg *protoregistry.Files, mute bool) (*Protos, error) {
	if fileReg == nil {
		fileReg = protoregistry.GlobalFiles
	}
	res := &Protos{
		protoc:  protoc,
		dir:     dir,
		makeFS:  makeFS,
		fileReg: fileReg,
		mute:    mute,
	}
	if err := res.registerFiles(); err != nil {
		return nil, err
	}
	return res, nil
}

// ProtoBytes converts a JSON representation of the specified message into protobuf bytes.
func (p *Protos) ProtoBytes(message protoreflect.FullName, fromJSON string) ([]byte, error) {
	md, err := p.messageDescriptor(message)
	if err != nil {
		return nil, err
	}
	dm := dynamicpb.NewMessage(md)
	if err = protojson.Unmarshal([]byte(fromJSON), dm); err != nil {
		return nil, err
	}
	res, err := proto.Marshal(dm)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// PrinterFor returns a function to friendly-print protobuf-encoded messages represented by om.
//
// TODO: default printer (when template is not deifined, prints all props+values)
func (p *Protos) PrinterFor(om *config.OutMessage) (func([]byte) (string, error), error) {
	md, err := p.messageDescriptor(om.Name)
	if err != nil {
		return nil, err
	}
	tplParamToFieldDescs, err := newTplParamToFieldDescs(md, om)
	if err != nil {
		return nil, err
	}
	mt := dynamicpb.NewMessageType(md)
	res := func(b []byte) (string, error) {
		rm := mt.New()
		m := rm.Interface()
		if err := proto.Unmarshal(b, m); err != nil {
			return "", err
		}
		var buf bytes.Buffer
		tplArgs := tplParamToFieldDescs.tplArgs(rm)
		if err := om.Template.Execute(&buf, tplArgs); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return res, nil
}

func (p *Protos) messageDescriptor(message protoreflect.FullName) (protoreflect.MessageDescriptor, error) {
	d, err := p.fileReg.FindDescriptorByName(message)
	if err != nil {
		return nil, err
	}
	res, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message descriptor")
	}
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

// TODO: absctract for testing
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
