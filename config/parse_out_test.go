package config

import (
	"strings"
	"testing"

	"github.com/m18/cpb/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestOutMessageParseAlias(t *testing.T) {
	tests := []struct {
		alias         string
		expectedAlias string
		err           bool
	}{
		{alias: "", err: true},
		{alias: "foo bar", err: true},
		{alias: "foo!", err: true},
		{alias: " foo", err: true},
		{
			alias:         "foo",
			expectedAlias: "foo",
		},
		{
			alias:         "foobar",
			expectedAlias: "foobar",
		},
		{
			alias:         "fooBar",
			expectedAlias: "fooBar",
		},
	}
	p := newOutMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.alias, func(t *testing.T) {
			t.Parallel()
			alias, err := p.parseAlias(test.alias)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if alias != test.expectedAlias {
				t.Fatalf("expected %q but got %q", test.expectedAlias, alias)
			}
		})
	}
}

func TestOutMessageParseTemplate(t *testing.T) {
	tests := []struct {
		tpl           string
		expectedProps map[string]struct{}
		err           bool
	}{
		{tpl: "", expectedProps: map[string]struct{}{}},
		{tpl: " ", expectedProps: map[string]struct{}{}},
		{
			tpl:           `foo - $foo, bar: $bar`,
			expectedProps: map[string]struct{}{"foo": {}, "bar": {}},
		},
		{
			tpl:           `foo - $foo!, bar: $bar???`,
			expectedProps: map[string]struct{}{"foo": {}, "bar": {}},
		},
		{
			tpl:           `foo - \$foo, bar: $bar`,
			expectedProps: map[string]struct{}{"bar": {}},
		},
		{
			tpl:           `foo - \$foo, bar: bar`,
			expectedProps: map[string]struct{}{},
		},
		{
			tpl:           `$foo}`,
			expectedProps: map[string]struct{}{"foo": {}},
		},
		{
			tpl:           `{{"{"}}}$foo}`, // surround output with "{" and "}"
			expectedProps: map[string]struct{}{"foo": {}},
		},
		{
			tpl:           `{ $foo}`, // no need for escaping if there is space after "{"
			expectedProps: map[string]struct{}{"foo": {}},
		},
		{
			tpl:           `(foobar: $foo.bar)`,
			expectedProps: map[string]struct{}{"foo.bar": {}},
		},
		{
			tpl:           `(foobar: \"$foo.bar\", \n foobarbaz = $foo.bar.baz)`,
			expectedProps: map[string]struct{}{"foo.bar": {}, "foo.bar.baz": {}},
		},
		{
			tpl: `{$foo}`, // this "{" needs escaping
			err: true,
		},
	}
	p := newOutMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.tpl, func(t *testing.T) {
			t.Parallel()
			tpl, props, err := p.parseTemplate("foo", test.tpl)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if tpl == nil {
				t.Fatalf("expected template to not be nil but it was")
			}
			if !check.StringSetsAreEqual(props, test.expectedProps) {
				t.Fatalf("expected %v but got %v", test.expectedProps, props)
			}
		})
	}
}

func TestOutMessageParseMessage(t *testing.T) {
	makeomc := func(cfg string) (string, *outMessageConfig) {
		var raw rawConfig
		if err := raw.from(strings.NewReader(cfg)); err != nil {
			t.Fatal(err)
		}
		for rawAlias, omc := range raw.Messages.Out {
			return rawAlias, omc
		}
		return "", nil
	}
	validConfig := `{
		"messages": {
			"out": {
				"foo": {
					"name": "proto.Foo",
					"template": "Hello, $world!"
				}
			}
		}
	}`
	validAlias, validomc := makeomc(validConfig)
	tests := []struct {
		desc          string
		rawAlias      string
		omc           outMessageConfig // cannot be nil
		expectedAlias string
		expectedName  protoreflect.FullName
		err           bool
	}{
		{
			desc:          "valid input",
			rawAlias:      validAlias,
			omc:           *validomc,
			expectedAlias: "foo",
			expectedName:  "proto.Foo",
		},
		{
			desc:          "empty message config",
			rawAlias:      "foo",
			omc:           outMessageConfig{},
			expectedAlias: "foo",
		},
		{
			desc:     "empty alias",
			rawAlias: "",
			omc:      *validomc,
			err:      true,
		},
		{
			desc:     "invalid alias",
			rawAlias: "!",
			omc:      *validomc,
			err:      true,
		},
		{
			desc:     "invalid template",
			rawAlias: validAlias,
			omc:      outMessageConfig{Template: "{{}}"},
			err:      true,
		},
	}
	p := newOutMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			om, err := p.parseMessage(test.rawAlias, &test.omc)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			if om == nil {
				t.Fatalf("expected output message to not be nil but it was")
			}
			if om.Alias != test.expectedAlias {
				t.Fatalf("expected alias to be %q but it was %q", test.expectedAlias, om.Alias)
			}
			if om.Name != test.expectedName {
				t.Fatalf("expected name to be %q but it was %q", test.expectedName, om.Name)
			}
		})
	}
}

func TestOutMessageParse(t *testing.T) {
	makeomcs := func(cfg string) map[string]*outMessageConfig {
		var raw rawConfig
		if err := raw.from(strings.NewReader(cfg)); err != nil {
			t.Fatal(err)
		}
		return raw.Messages.Out
	}
	validConfig := `{
		"messages": {
			"out": {
				"foo": {
					"name": "proto.Foo",
					"template": "Hello, $world!"
				},
				"bar": {
					"name": "proto.Bar",
					"template": "Hi, $cosmos."
				}
			}
		}
	}`
	duplicateAliasConfig := `{
		"messages": {
			"out": {
				"foo": {
					"name": "proto.Foo",
					"template": "Hello, $world!"
				},
				"foo": {
					"name": "proto.Bar",
					"template": "Hi, $cosmos."
				}
			}
		}
	}`
	tests := []struct {
		desc                    string
		omcs                    map[string]*outMessageConfig
		expectedAliasesAndNames map[string]protoreflect.FullName
		err                     bool
	}{
		{
			desc: "valid config",
			omcs: makeomcs(validConfig),
			expectedAliasesAndNames: map[string]protoreflect.FullName{
				"foo": "proto.Foo",
				"bar": "proto.Bar",
			},
		},
		{
			desc: "duplicate alias, last one takes precedence",
			omcs: makeomcs(duplicateAliasConfig),
			expectedAliasesAndNames: map[string]protoreflect.FullName{
				"foo": "proto.Bar",
			},
		},
		{
			desc:                    "empty config",
			omcs:                    nil,
			expectedAliasesAndNames: nil,
		},
	}
	p := newOutMessageParser()
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			oms, err := p.parse(test.omcs)
			if err == nil == test.err {
				t.Fatalf("expected %t but didn't get it: %v", test.err, err)
			}
			if test.err {
				return
			}
			l, el := len(oms), len(test.expectedAliasesAndNames)
			if l != el {
				t.Fatalf("expected number of messages to be %d but it was %d", el, l)
			}
			for expectedAlias, expectedName := range test.expectedAliasesAndNames {
				om, ok := oms[expectedAlias]
				if !ok {
					t.Fatalf("expected %t but didn't get it: %v", test.err, err)
				}
				if om.Name != expectedName {
					t.Fatalf("expected name to be %q but it was %q", expectedName, om.Name)
				}
			}
		})
	}
}
