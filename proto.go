package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"

	// import well-known types
	// https://pkg.go.dev/google.golang.org/protobuf/types/known
	_ "google.golang.org/protobuf/types/known/timestamppb"
)

type protos struct {
	dir        string
	protoc     string
	dsFileName string
	fileReg    *protoregistry.Files
}

// newProtos returns a new protos for dir
func newProtos(dir string) (*protos, error) {
	res := &protos{
		dir:        dir,
		protoc:     "protoc",
		dsFileName: ".tmp_proto.ds",
		fileReg:    protoregistry.GlobalFiles,
	}
	if err := res.regFiles(); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *protos) protoBytes(name protoreflect.FullName, fromJSON string) ([]byte, error) {
	d, err := p.fileReg.FindDescriptorByName(name)
	if err != nil {
		return nil, err
	}
	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message descriptor")
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

func (p *protos) fromProtoBytes(name protoreflect.FullName, b []byte) (*dynamicpb.Message, error) {
	// extract
	d, err := p.fileReg.FindDescriptorByName(name)
	if err != nil {
		return nil, err
	}
	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message descriptor")
	}
	// ^ extract

	mt := dynamicpb.NewMessageType(md)
	rm := mt.New()
	m := rm.Interface()
	if err := proto.Unmarshal(b, m); err != nil {
		return nil, err
	}

	blah := md.Fields().ByName("blah")
	rm.Get(blah).Interface()

	// dm.ProtoReflect().
	return nil, nil
}

func (p *protos) regFiles() error {
	files, err := p.protoFiles()
	if err != nil {
		return err
	}
	if err = p.createDescSet(files); err != nil {
		return err
	}
	defer func() {
		if err = p.removeDescSet(); err != nil {
			// TODO: proper logging
			log.Println(err)
		}
	}()
	if err = p.regDescSet(); err != nil {
		return err
	}
	return nil
}

func (p *protos) protoFiles() ([]string, error) {
	return []string{path.Join(p.dir, "sample.proto")}, nil // TODO: get files from dir
}

func (p *protos) createDescSet(files []string) error {
	args := append(
		[]string{"-I", p.dir, "--descriptor_set_out", p.dsFileName}, // TODO: check with nested dirs
		files...,
	)
	cmd := exec.Command(p.protoc, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (p *protos) removeDescSet() error {
	return os.Remove(p.dsFileName)
}

func (p *protos) regDescSet() error {
	// TODO: go 1.16
	dsData, err := ioutil.ReadFile(p.dsFileName)
	if err != nil {
		return err
	}
	fds := &descriptorpb.FileDescriptorSet{}
	if err = proto.Unmarshal(dsData, fds); err != nil {
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
