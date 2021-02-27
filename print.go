package main

import (
	"bytes"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// TODO: overall -- split into smaller bits for readability & testability
func printer(p *protos, om *OutMessage) (func([]byte) (string, error), error) {
	d, err := p.fileReg.FindDescriptorByName(om.Name)
	if err != nil {
		return nil, err
	}
	md, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil, fmt.Errorf("not a message descriptor")
	}

	tplParamToFieldDescs := map[string][]protoreflect.FieldDescriptor{}

	for dotProps := range om.props {
		props := strings.Split(dotProps, ".")
		fds := make([]protoreflect.FieldDescriptor, 0, len(props))
		mdCurr := md
		for _, prop := range props {
			if mdCurr == nil {
				return nil, fmt.Errorf("invalid property: %s", dotProps)
			}
			fd := mdCurr.Fields().ByName(protoreflect.Name(prop))
			if fd == nil {
				return nil, fmt.Errorf("invalid property name: %s (%s)", prop, dotProps)
			}
			fds = append(fds, fd)
			mdCurr = fd.Message()
		}
		tplParamToFieldDescs[propToTplParam(dotProps)] = fds
	}

	mt := dynamicpb.NewMessageType(md)
	res := func(b []byte) (string, error) {
		rm := mt.New()
		m := rm.Interface()
		if err := proto.Unmarshal(b, m); err != nil {
			return "", err
		}

		tplArgs := map[string]interface{}{}
		for tplParam, fds := range tplParamToFieldDescs {
			v := protoreflect.ValueOf(rm) // just to play nice with the for loop
			for _, fd := range fds {
				v = v.Message().Get(fd)
			}
			tplArgs[tplParam] = v.Interface()
		}

		var buf bytes.Buffer
		if err := om.template.Execute(&buf, tplArgs); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return res, nil
}

// TODO: handle maps and lists
//        in: (..., [123,456]), insert into samples(id, nam, dat) values(1, 'blah', $p(12, \"blah2\", 2010, \"uno\", [123,456]))
//       out: $phones[0] $phones[*] (all?)
//
//       rm.Get(fd).List().Get(0).Message().Get(fd2)
func rangeRec(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
	if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() {
		v.Message().Range(rangeRec)
	}
	return true
}
