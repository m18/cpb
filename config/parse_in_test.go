package config

import (
	"strings"
	"testing"

	"github.com/m18/cpb/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestInMessageParseAlias(t *testing.T) {
	tests := []struct {
		aliasWithParams string
		expectedAlias   string
		expectedParams  string
		err             bool
	}{
		{aliasWithParams: "", err: true},
		{aliasWithParams: " ", err: true},
		{aliasWithParams: "  ", err: true},
		{aliasWithParams: "!()", err: true},
		{aliasWithParams: "p!()", err: true},
		{aliasWithParams: "p(id name)", err: true},
		{
			aliasWithParams: "p()",
			expectedAlias:   "p",
			expectedParams:  "",
		},
		{
			aliasWithParams: "fooBar()",
			expectedAlias:   "fooBar",
			expectedParams:  "",
		},
		{
			aliasWithParams: "p(id, name)",
			expectedAlias:   "p",
			expectedParams:  "id, name",
		},
		{
			aliasWithParams: " foo  (  id ,  name  )",
			expectedAlias:   "foo",
			expectedParams:  "  id ,  name  ",
		},
		{
			aliasWithParams: " foo  (id, id)",
			expectedAlias:   "foo",
			expectedParams:  "id, id",
		},
	}
	p := newInMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.aliasWithParams, func(t *testing.T) {
			t.Parallel()
			alias, params, err := p.parseAlias(test.aliasWithParams)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if alias != test.expectedAlias {
				t.Fatalf("expected %q but got %q", test.expectedAlias, alias)
			}
			if params != test.expectedParams {
				t.Fatalf("expected %q but got %q", test.expectedParams, params)
			}
		})
	}
}

func TestInMessageParseAliasParams(t *testing.T) {
	tests := []struct {
		aliasParams         string
		expectedParams      []string
		expectedParamLookup map[string]struct{}
		err                 bool
	}{
		{aliasParams: "id, id", err: true},
		{aliasParams: "  id ,  id  ", err: true},
		{
			aliasParams:         "",
			expectedParams:      []string{},
			expectedParamLookup: map[string]struct{}{},
		},
		{
			aliasParams:         " ",
			expectedParams:      []string{},
			expectedParamLookup: map[string]struct{}{},
		},
		{
			aliasParams:         "  ",
			expectedParams:      []string{},
			expectedParamLookup: map[string]struct{}{},
		},
		{
			aliasParams:         "id",
			expectedParams:      []string{"id"},
			expectedParamLookup: map[string]struct{}{"id": {}},
		},
		{
			aliasParams:         "id, name",
			expectedParams:      []string{"id", "name"},
			expectedParamLookup: map[string]struct{}{"id": {}, "name": {}},
		},
		{
			aliasParams:         "  id  ,  name   ",
			expectedParams:      []string{"id", "name"},
			expectedParamLookup: map[string]struct{}{"name": {}, "id": {}},
		},
	}
	p := newInMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.aliasParams, func(t *testing.T) {
			t.Parallel()
			params, paramLookup, err := p.parseAliasParams("", test.aliasParams)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if !check.StringSlicesAreEqual(params, test.expectedParams) {
				t.Fatalf("expected %v but got %v", test.expectedParams, params)
			}
			if !check.StringSetsAreEqual(paramLookup, test.expectedParamLookup) {
				t.Fatalf("expected %v to contain %v but it did not", paramLookup, test.expectedParamLookup)
			}
		})
	}
}

func TestInMessageParseTemplate(t *testing.T) {
	makeInMessageTpl := func(cfg string) interface{} {
		var raw rawConfig
		if err := raw.from(strings.NewReader(cfg)); err != nil {
			t.Fatal(err)
		}
		for _, v := range raw.Messages.In {
			return v.Template // return the first message template
		}
		return nil
	}

	tplm := makeInMessageTpl(`{"messages": {"in": {"_": {"template":{"one":"$one","two":"$two"}}}}}`)
	tplmNoVars := makeInMessageTpl(`{"messages": {"in": {"_": {"template":{"one":"one","two":"$$two"}}}}}`)

	tests := []struct {
		desc        string
		tplm        interface{}
		paramLookup map[string]struct{}
		err         bool
	}{
		{
			desc:        "nil template, nil paramLookup",
			tplm:        nil,
			paramLookup: nil,
			err:         false,
		},
		{
			desc:        "nil template with paramLookup",
			tplm:        nil,
			paramLookup: map[string]struct{}{"one": {}, "two": {}},
			err:         false,
		},
		{
			desc:        "template with vars, matching paramLookup",
			tplm:        tplm,
			paramLookup: map[string]struct{}{"one": {}, "two": {}},
			err:         false,
		},
		{
			desc:        "template with vars, extra params in paramLookup",
			tplm:        tplm,
			paramLookup: map[string]struct{}{"one": {}, "two": {}, "threee": {}},
			err:         false,
		},
		{
			desc:        "template with no vars, with no paramLookup",
			tplm:        tplmNoVars,
			paramLookup: nil,
			err:         false,
		},
		{
			desc:        "template with no vars, with paramLookup",
			tplm:        tplmNoVars,
			paramLookup: map[string]struct{}{"one": {}, "two": {}, "threee": {}},
			err:         false,
		},
		{
			desc:        "template with vars, with no paramLookup",
			tplm:        tplm,
			paramLookup: nil,
			err:         true,
		},
		{
			desc:        "template with vars, non-matching paramLookup",
			tplm:        tplm,
			paramLookup: map[string]struct{}{"one": {}},
			err:         true,
		},
	}
	p := newInMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			tpl, err := p.parseTemplate("foo", test.tplm, test.paramLookup)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if !test.err && tpl == nil {
				t.Fatalf("expected tpl to not be nil but it was")
			}
		})
	}
}

func TestInMessageParseMessage(t *testing.T) {
	makeimc := func(cfg string) (string, *inMessageConfig) {
		var raw rawConfig
		if err := raw.from(strings.NewReader(cfg)); err != nil {
			t.Fatal(err)
		}
		for aliasWithParams, imc := range raw.Messages.In {
			return aliasWithParams, imc
		}
		return "", nil
	}
	validConfig := `{
		"messages": {
			"in": {
				"foo(id, name)": {
					"name": "proto.Foo",
					"template": {
						"id": "$id",
						"name": "$name"						
					}
				}
			}
		}
	}`
	invalidConfig := `{
		"messages": {
			"in": {
				"foo(name)": {
					"name": "proto.Foo",
					"template": {
						"id": "$id",
						"name": "$name"						
					}
				}
			}
		}
	}`
	validAliasWithParams, validimc := makeimc(validConfig)
	invalidAliasWithParams, invalidimc := makeimc(invalidConfig)
	tests := []struct {
		desc            string
		aliasWithParams string
		imc             inMessageConfig // cannot be nil
		expectedAlias   string
		expectedParams  []string
		expectedName    protoreflect.FullName
		err             bool
	}{
		{
			desc:            "valid input",
			aliasWithParams: validAliasWithParams,
			imc:             *validimc,
			expectedAlias:   "foo",
			expectedParams:  []string{"id", "name"},
			expectedName:    "proto.Foo",
		},
		{
			desc:            "empty message config",
			aliasWithParams: validAliasWithParams,
			imc:             inMessageConfig{},
			expectedAlias:   "foo",
			expectedParams:  []string{"id", "name"},
		},
		{
			desc:            "empty alias",
			aliasWithParams: "",
			imc:             *validimc,
			err:             true,
		},
		{
			desc:            "invalid alias",
			aliasWithParams: "!()",
			imc:             *validimc,
			err:             true,
		},
		{
			desc:            "invalid input",
			aliasWithParams: invalidAliasWithParams,
			imc:             *invalidimc,
			err:             true,
		},
	}
	p := newInMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			im, err := p.parseMessage(test.aliasWithParams, &test.imc)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if im == nil {
				t.Fatalf("expected output message to not be nil but it was")
			}
			if im.Alias != test.expectedAlias {
				t.Fatalf("expected alias to be %q but it was %q", test.expectedAlias, im.Alias)
			}
			if !check.StringSlicesAreEqual(im.params, test.expectedParams) {
				t.Fatalf("expected params to be %v but they were %v", test.expectedParams, im.params)
			}
			if im.Name != test.expectedName {
				t.Fatalf("expected name to be %q but it was %q", test.expectedName, im.Name)
			}
		})
	}
}

func TestInMessageParse(t *testing.T) {
	makeimcs := func(cfg string) map[string]*inMessageConfig {
		var raw rawConfig
		if err := raw.from(strings.NewReader(cfg)); err != nil {
			t.Fatal(err)
		}
		return raw.Messages.In
	}
	validConfig := `{
		"messages": {
			"in": {
				"foo(id, name)": {
					"name": "proto.Foo",
					"template": {
						"id": "$id",
						"name": "$name"
					}
				},
				"bar(type, color)": {
					"name": "proto.Bar",
					"template": {
						"type": "$type",
						"color": "$color"
					}
				}
			}
		}
	}`
	duplicateAliasConfig := `{
		"messages": {
			"in": {
				"foo(id, name)": {
					"name": "proto.Foo",
					"template": {
						"id": "$id",
						"name": "$name"
					}
				},
				"foo(type, color)": {
					"name": "proto.Foo",
					"template": {
						"type": "$type",
						"color": "$color"
					}
				}
			}
		}
	}`
	undetectableDuplicateAliasConfig := `{
		"messages": {
			"in": {
				"foo(id, name)": {
					"name": "proto.Foo",
					"template": {
						"id": "$id",
						"name": "$name"
					}
				},
				"foo(id, name)": {
					"name": "proto.Bar",
					"template": {
						"type": "$id",
						"color": "$name"
					}
				}
			}
		}
	}`
	tests := []struct {
		desc                    string
		imcs                    map[string]*inMessageConfig
		expectedAliasesAndNames map[string]protoreflect.FullName
		err                     bool
	}{
		{
			desc: "valid config",
			imcs: makeimcs(validConfig),
			expectedAliasesAndNames: map[string]protoreflect.FullName{
				"foo": "proto.Foo",
				"bar": "proto.Bar",
			},
		},
		{
			desc: "duplicate alias",
			imcs: makeimcs(duplicateAliasConfig),
			err:  true,
		},
		{
			desc: "undetactable duplicate alias, last one takes precedence",
			imcs: makeimcs(undetectableDuplicateAliasConfig),
			expectedAliasesAndNames: map[string]protoreflect.FullName{
				"foo": "proto.Bar",
			},
		},
		{
			desc:                    "empty config",
			imcs:                    nil,
			expectedAliasesAndNames: nil,
		},
	}
	p := newInMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			ims, err := p.parse(test.imcs)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			l, el := len(ims), len(test.expectedAliasesAndNames)
			if l != el {
				t.Fatalf("expected number of messages to be %d but it was %d", el, l)
			}
			for expectedAlias, expectedName := range test.expectedAliasesAndNames {
				im, ok := ims[expectedAlias]
				if !ok {
					t.Fatalf("expected alias %s to be present it was not", expectedAlias)
				}
				if im.Name != expectedName {
					t.Fatalf("expected name to be %q but it was %q", expectedName, im.Name)
				}
			}
		})
	}
}
