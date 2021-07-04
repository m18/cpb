package protos

import (
	"strings"
	"testing"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/internal/testcheck"
	"github.com/m18/cpb/internal/tmpl"
	"github.com/m18/eq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func TestTplParamToFieldDescsNew(t *testing.T) {
	md, err := barLiteMessageDescriptor()
	testcheck.FatalIf(t, err)
	tests := []struct {
		desc string
		md   protoreflect.MessageDescriptor
		om   *config.OutMessage
		err  bool
	}{
		{
			desc: "empty props",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{},
			},
		},
		{
			desc: "one prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"text": {},
				},
			},
		},
		{
			desc: "one nested prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"nested.name": {},
				},
			},
		},
		{
			desc: "mutiple props",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":          {},
					"text":        {},
					"nested.name": {},
				},
			},
		},
		{
			desc: "one non-existent prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"nonexistent": {},
				},
			},
			err: true,
		},
		{
			desc: "non-existent prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":          {},
					"nonexistent": {},
				},
			},
			err: true,
		},
		{
			desc: "one fully non-existent nested prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"non.existent": {},
				},
			},
			err: true,
		},
		{
			desc: "fully non-existent nested prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":           {},
					"non.existent": {},
				},
			},
			err: true,
		},
		{
			desc: "one partially non-existent nested prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"text.nonexistent": {},
				},
			},
			err: true,
		},
		{
			desc: "partially non-existent nested prop",
			md:   md,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":                      {},
					"nested.name.nonexistent": {},
				},
			},
			err: true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			tplParamToFieldDescs, err := newTplParamToFieldDescs(test.md, test.om)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if tplParamToFieldDescs == nil {
				t.Fatalf("expected tplParamToFieldDescs to not be nil but it was")
			}
			if len(tplParamToFieldDescs) != len(test.om.Props) {
				t.Fatalf("expected len(tplParamToFieldDescs) to be %d but it was %d", len(test.om.Props), len(tplParamToFieldDescs))
			}
			for prop := range test.om.Props {
				tplParam := tmpl.PropToTemplateParam(prop)
				expectedCount := strings.Count(prop, ".") + 1
				actualCount := len(tplParamToFieldDescs[tplParam])
				if actualCount != expectedCount {
					t.Fatalf("expected count to be %d but it was %d", expectedCount, actualCount)
				}
			}
		})
	}
}

func TestTplParamToFieldDescsTplArgs(t *testing.T) {
	md, err := barLiteMessageDescriptor()
	testcheck.FatalIf(t, err)
	tests := []struct {
		desc     string
		json     string
		om       *config.OutMessage
		expected map[string]interface{}
	}{
		{
			desc: "more props available than used",
			json: `{"id": 1, "text": "foo"}`,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"text": {},
				},
			},
			expected: map[string]interface{}{
				"text": "foo",
			},
		},
		{
			desc: "all available props used",
			json: `{"id": 1, "text": "foo"}`,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":   {},
					"text": {},
				},
			},
			expected: map[string]interface{}{
				"id":   int32(1),
				"text": "foo",
			},
		},
		{
			desc: "all available props used, nested",
			json: `{"id": 1, "text": "foo", "nested": {"name": "bar"}}`,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":          {},
					"text":        {},
					"nested.name": {},
				},
			},
			expected: map[string]interface{}{
				"id":                                    int32(1),
				"text":                                  "foo",
				tmpl.PropToTemplateParam("nested.name"): "bar",
			},
		},
		{
			desc: "fewer props available than used",
			json: `{"text": "foo"}`,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":   {},
					"text": {},
				},
			},
			expected: map[string]interface{}{
				"id":   int32(0),
				"text": "foo",
			},
		},
		{
			desc: "fewer props available than used, nested",
			json: `{"text": "foo"}`,
			om: &config.OutMessage{
				Props: map[string]struct{}{
					"id":          {},
					"text":        {},
					"nested.name": {},
				},
			},
			expected: map[string]interface{}{
				"id":                                    int32(0),
				"text":                                  "foo",
				tmpl.PropToTemplateParam("nested.name"): "",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			tplParamToFieldDescs, err := newTplParamToFieldDescs(md, test.om)
			testcheck.FatalIf(t, err)
			dm := dynamicpb.NewMessage(md)
			testcheck.FatalIf(t, protojson.Unmarshal([]byte(test.json), dm))
			args := tplParamToFieldDescs.tplArgs(dm)
			if !eq.StringToSimpleTypeMaps(args, test.expected) {
				t.Fatalf("expected %v but got %v", test.expected, args)
			}
		})
	}
}
