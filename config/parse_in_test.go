package config

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m18/cpb/check"
)

func TestInMessageParseAlias(t *testing.T) {
	tests := []struct {
		aliasWithParams string
		expectedAlias   string
		expectedParams  string
		err             bool
	}{
		{aliasWithParams: "", err: true},
		{aliasWithParams: "!()", err: true},
		{aliasWithParams: "p!()", err: true},
		{aliasWithParams: "p(id name)", err: true},
		{
			aliasWithParams: "p()",
			expectedAlias:   "p",
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
				t.Fatalf("expected %t but didn't get it", test.err)
			}
			if alias != test.expectedAlias {
				t.Fatalf("expected %s but got %s", test.expectedAlias, alias)
			}
			if params != test.expectedParams {
				t.Fatalf("expected %s but got %s", test.expectedParams, params)
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
				t.Fatalf("expected %t but didn't get it", test.err)
			}
			if !check.StringSlicesAreEqual(params, test.expectedParams) {
				t.Fatalf("expected %s but got %s", test.expectedParams, params)
			}
			if !check.StringSetsAreEqual(paramLookup, test.expectedParamLookup) {
				t.Fatalf("expected %v to contain %v but it did not", paramLookup, test.expectedParamLookup)
			}
		})
	}
}

func TestInMessageParseTemplate(t *testing.T) {
	makeInMessageTpl := func(tpl string) (interface{}, error) {
		var raw rawConfig
		if err := raw.from(strings.NewReader(tpl)); err != nil {
			return nil, err
		}
		for _, v := range raw.Messages.In {
			return v.Template, nil // return the first message template
		}
		return nil, fmt.Errorf("no in-message templates defined")
	}

	tplm, err := makeInMessageTpl(`{"messages": {"in": {"_": {"template":{"one":"$one","two":"$two"}}}}}`)
	if err != nil {
		t.Fatalf("invalid template string")
	}
	tplmNoVars, err := makeInMessageTpl(`{"messages": {"in": {"_": {"template":{"one":"one","two":"$$two"}}}}}`)
	if err != nil {
		t.Fatalf("invalid template string")
	}

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
