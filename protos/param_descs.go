package protos

import (
	"fmt"
	"strings"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/tmpl"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type tplParamToFieldDescs map[string][]protoreflect.FieldDescriptor

// TODO: handle maps and lists
//        in: (..., [123,456]), insert into samples(id, nam, dat) values(1, 'blah', $p(12, \"blah2\", 2010, \"uno\", [123,456]))
//       out: $phones[0] $phones[*] (all?)
//
//       rm.Get(fd).List().Get(0).Message().Get(fd2)
// func rangeRec(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
// 	if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() {
// 		v.Message().Range(rangeRec)
// 	}
// 	return true
// }
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
