package protos

import (
	"fmt"
	"strings"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/tmpl"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type tplParamToFieldDescs map[string][]protoreflect.FieldDescriptor

func newTplParamToFieldDescs(md protoreflect.MessageDescriptor, om *config.OutMessage) (tplParamToFieldDescs, error) {
	res := tplParamToFieldDescs{}
	for dotProps := range om.Props {
		props := strings.Split(dotProps, ".")
		fds := make([]protoreflect.FieldDescriptor, 0, len(props))
		currmd := md
		for _, prop := range props {
			if currmd == nil {
				return nil, fmt.Errorf("invalid property: %s", dotProps)
			}
			fd := currmd.Fields().ByName(protoreflect.Name(prop))
			if fd == nil {
				return nil, fmt.Errorf("invalid property name: %s (%s)", prop, dotProps)
			}
			fds = append(fds, fd)
			currmd = fd.Message()
		}
		res[tmpl.PropToTemplateParam(dotProps)] = fds
	}
	return res, nil
}

func (m tplParamToFieldDescs) tplArgs(rm protoreflect.Message) map[string]interface{} {
	res := map[string]interface{}{}
	for tplParam, fds := range m {
		v := protoreflect.ValueOf(rm)
		for _, fd := range fds {
			v = v.Message().Get(fd)
		}
		res[tplParam] = v.Interface()
	}
	return res
}
