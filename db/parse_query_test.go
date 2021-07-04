package db

import (
	"fmt"
	"testing"

	"github.com/m18/cpb/internal/testcheck"
	"github.com/m18/cpb/internal/testconfig"
	"github.com/m18/cpb/internal/testprotos"
	"github.com/m18/eq"
)

func TestQueryParserNormalizeInMessageArgs(t *testing.T) {
	tests := []struct {
		args         []string
		expectedArgs []string
	}{
		{
			args:         []string{},
			expectedArgs: []string{},
		},
		{
			args:         []string{""},
			expectedArgs: []string{""},
		},
		{
			args:         []string{"", ""},
			expectedArgs: []string{"", ""},
		},
		{
			args:         []string{"'foo'", "1"},
			expectedArgs: []string{"\"foo\"", "1"},
		},
		{
			args:         []string{"0", "1"},
			expectedArgs: []string{"0", "1"},
		},
		{
			args:         []string{"'foo'", "'bar'"},
			expectedArgs: []string{"\"foo\"", "\"bar\""},
		},
		{
			args:         []string{"'foo bar'", "' bar baz '"},
			expectedArgs: []string{"\"foo bar\"", "\" bar baz \""},
		},
		{
			args:         []string{"'o\\'foo o\\'bar'", "' bar o\\'baz '"},
			expectedArgs: []string{"\"o'foo o'bar\"", "\" bar o'baz \""},
		},
	}
	qp := newQueryParser(DriverPostgres, nil, nil, nil, false)
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprint(test.args), func(t *testing.T) {
			t.Parallel()
			args := qp.normalizeInMessageArgs(test.args)
			if !eq.StringSlices(args, test.expectedArgs) {
				t.Fatalf("expected %v but got %v", test.expectedArgs, args)
			}
		})
	}

}

func TestQueryParserParse(t *testing.T) {
	p, err := testprotos.MakeProtosLite()
	testcheck.FatalIf(t, err)
	tests := []struct {
		desc          string
		driver        string
		query         string
		isPlainQuery  bool
		expectedQuery string
		err           bool
	}{
		{
			desc:          "valid input, plain",
			driver:        DriverPostgres,
			query:         "select * from test;",
			isPlainQuery:  true,
			expectedQuery: "select * from test;",
		},
		{
			desc:          "valid input",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col from test where bar_col = $bar(2, 'two');",
			expectedQuery: "select foo_col from test where bar_col = $1;",
		},
		{
			desc:          "valid, extra spaces",
			driver:        DriverPostgres,
			query:         " select  $foo:foo_col   from test where  bar_col  = $bar( 2 , 'two') ;  ",
			expectedQuery: " select  foo_col   from test where  bar_col  = $1 ;  ",
		},
		{
			desc:   `invalid, unknown "in" alias`,
			driver: DriverPostgres,
			query:  "select $foo:foo_col from test where bar_col = $unknown(2, 'two');",
			err:    true,
		},
		{
			desc:   `invalid, unknown "out" alias`,
			driver: DriverPostgres,
			query:  "select $unknown:foo_col from test where bar_col = $bar(2, 'two');",
			err:    true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			cfg, err := testconfig.MakeTestConfigLite(test.driver)
			testcheck.FatalIf(t, err)
			qp := newQueryParser(cfg.DB.Driver, p, cfg.InMessages, cfg.OutMessages, false)
			q, inMessageArgs, outMessageStringers, err := qp.parse(test.query)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if q != test.expectedQuery {
				t.Fatalf("expected query to be %q but it was %q", test.expectedQuery, q)
			}
			if inMessageArgs == nil && !test.isPlainQuery {
				t.Fatalf("expected inMessageArgs to not be nil but it was")
			}
			if outMessageStringers == nil {
				t.Fatalf("expected outMessageStringers to not be nil but it was")
			}
		})
	}
}

func TestQueryParserParseInMessageArgs(t *testing.T) {
	p, err := testprotos.MakeProtosLite()
	testcheck.FatalIf(t, err)
	tests := []struct {
		desc             string
		driver           string
		query            string
		expectedQuery    string
		expectedArgCount int
		err              bool
	}{
		{
			desc:             "valid, no args",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = 100",
			expectedQuery:    "select * from test where foo_col = 100",
			expectedArgCount: 0,
		},
		{
			desc:             "valid, malformed arg",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1",
			expectedQuery:    "select * from test where foo_col = $foo(1",
			expectedArgCount: 0,
		},
		{
			desc:             "valid, single arg",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single arg, type coersion - '1' -> 1",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo('1', 'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single arg, extra spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col   =  $foo(  1 , 'one' , true )  ",
			expectedQuery:    "select * from test where foo_col   =  $1  ",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single arg, no spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col=$foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col=$1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single arg, no left space",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col= $foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col= $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single arg, no right space",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col =$foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col =$1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, multiple args",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'one', true) and bar_col = $bar(2, 'two')",
			expectedQuery:    "select * from test where foo_col = $1 and bar_col = $2",
			expectedArgCount: 2,
		},
		{
			desc:             "valid, multiple args, mixed spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col=$foo(   1   ,  'one'   , true )  and   bar_col =$bar( 2, 'two' ) ",
			expectedQuery:    "select * from test where foo_col=$1  and   bar_col =$2 ",
			expectedArgCount: 2,
		},
		{
			desc:             "valid, single arg, escaped quote",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'o\\'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, multiple args, escaped quotes",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'o\\'one', true) and bar_col = $bar(2, 'o\\'two')",
			expectedQuery:    "select * from test where foo_col = $1 and bar_col = $2",
			expectedArgCount: 2,
		},
		{
			desc:             "valid, JSON-escaped double quote",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'one \\\" two', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, empty",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $empty()",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:   "invalid, JSON-unescaped double quote",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one \" two', false)",
			err:    true,
		},
		{
			desc:   "invalid, single arg, wrong sub-arg type (string instead of int32)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo('not a number', 'one', false)",
			err:    true,
		},
		{
			desc:   "invalid, single arg, wrong sub-arg type (string instead of bool)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one', 'false')",
			err:    true,
		},
		{
			desc:   "invalid, single arg, wrong sub-arg type (float instead of int32)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1.1, 'one', false)",
			err:    true,
		},
		{
			desc:   "invalid, single arg, wrong sub-arg count",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1)",
			err:    true,
		},
		{
			desc:   "invalid, multiple args, wrong sub-arg type",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one', false) and bar_col = $bar('not a number', 'two')",
			err:    true,
		},
		{
			desc:   "invalid, multiple args, wrong sub-arg count",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one', false) and bar_col = $bar(2)",
			err:    true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			cfg, err := testconfig.MakeTestConfigLite(test.driver)
			testcheck.FatalIf(t, err)
			qp := newQueryParser(cfg.DB.Driver, p, cfg.InMessages, nil, false)
			q, args, err := qp.parseInMessageArgs(test.query)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if q != test.expectedQuery {
				t.Fatalf("expected query to be %q but it was %q", test.expectedQuery, q)
			}
			if len(args) != test.expectedArgCount {
				t.Fatalf("expected arg count to be %d but it was %d", test.expectedArgCount, len(args))
			}
		})
	}
}

func TestQueryParserParseOutMessageArgs(t *testing.T) {
	p, err := testprotos.MakeProtosLite()
	testcheck.FatalIf(t, err)
	tests := []struct {
		desc                 string
		driver               string
		autoMapOutMessages   bool
		query                string
		expectedQuery        string
		expectedStringerKeys map[string]struct{}
		err                  bool
	}{
		{
			desc:          "valid, no args",
			driver:        DriverPostgres,
			query:         "select * from test",
			expectedQuery: "select * from test",
		},
		{
			desc:          "valid, no args, $",
			driver:        DriverPostgres,
			query:         "select $foo from test",
			expectedQuery: "select $foo from test",
		},
		{
			desc:          "valid, no args, $:",
			driver:        DriverPostgres,
			query:         "select $foo: from test",
			expectedQuery: "select $foo: from test",
		},
		{
			desc:          "valid, no args, $:space",
			driver:        DriverPostgres,
			query:         "select $foo: blah from test",
			expectedQuery: "select $foo: blah from test",
		},
		{
			desc:          "valid, no args, extra spaces",
			driver:        DriverPostgres,
			query:         "select  *  from    test ",
			expectedQuery: "select  *  from    test ",
		},
		{
			desc:          "valid, single arg",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col from test",
			expectedQuery: "select foo_col from test",
			expectedStringerKeys: map[string]struct{}{
				"foo_col": {},
			},
		},
		{
			desc:          "valid, single arg, extra spaces",
			driver:        DriverPostgres,
			query:         "  select $foo:foo_col    from      test",
			expectedQuery: "  select foo_col    from      test",
			expectedStringerKeys: map[string]struct{}{
				"foo_col": {},
			},
		},
		{
			desc:          "valid, single arg with alias",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col as bar_col from test",
			expectedQuery: "select foo_col as bar_col from test",
			expectedStringerKeys: map[string]struct{}{
				"bar_col": {},
			},
		},
		{
			desc:          "valid, single arg with alias, extra spaces, capital characters",
			driver:        DriverPostgres,
			query:         " select  $foo:foo_Col    aS     bar_col  from test",
			expectedQuery: " select  foo_Col    aS     bar_col  from test",
			expectedStringerKeys: map[string]struct{}{
				"bar_col": {},
			},
		},
		{
			desc:          "valid, multiple args",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col, $bar:bar_col from test",
			expectedQuery: "select foo_col, bar_col from test",
			expectedStringerKeys: map[string]struct{}{
				"foo_col": {},
				"bar_col": {},
			},
		},
		{
			desc:          "valid, multiple args, extra spaces",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col  ,  $bar:bar_col    from  test",
			expectedQuery: "select foo_col  ,  bar_col    from  test",
			expectedStringerKeys: map[string]struct{}{
				"foo_col": {},
				"bar_col": {},
			},
		},
		{
			desc:          "valid, multiple args with aliases",
			driver:        DriverPostgres,
			query:         "select $foo:foo_col as baz_col, $bar:bar_col as qux_col from test",
			expectedQuery: "select foo_col as baz_col, bar_col as qux_col from test",
			expectedStringerKeys: map[string]struct{}{
				"baz_col": {},
				"qux_col": {},
			},
		},
		{
			desc:          "valid, multiple args with aliases, extra spaces, capital characters",
			driver:        DriverPostgres,
			query:         "select  $foo:foo_col AS  baz_cOl  ,  $bar:bar_Col   As  QuX_Col   FROM teSt ",
			expectedQuery: "select  foo_col AS  baz_cOl  ,  bar_Col   As  QuX_Col   FROM teSt ",
			expectedStringerKeys: map[string]struct{}{
				"baz_cOl": {},
				"QuX_Col": {},
			},
		},
		{
			desc:               "valid, auto-map, plain",
			driver:             DriverPostgres,
			query:              "select * from test",
			autoMapOutMessages: true,
			expectedQuery:      "select * from test",
			expectedStringerKeys: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			desc:               "valid, auto-map, single arg",
			driver:             DriverPostgres,
			query:              "select $foo:baz from test",
			autoMapOutMessages: true,
			expectedQuery:      "select baz from test",
			expectedStringerKeys: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
		{
			desc:               "valid, auto-map, single arg, override auto-mapped",
			driver:             DriverPostgres,
			query:              "select $bar:foo from test",
			autoMapOutMessages: true,
			expectedQuery:      "select foo from test",
			expectedStringerKeys: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			desc:               "valid, auto-map, single arg with alias, override auto-mapped",
			driver:             DriverPostgres,
			query:              "select $bar:baz as foo from test",
			autoMapOutMessages: true,
			expectedQuery:      "select baz as foo from test",
			expectedStringerKeys: map[string]struct{}{
				"foo": {},
				"bar": {},
			},
		},
		{
			desc:   "invalid, single arg, unknown alias",
			driver: DriverPostgres,
			query:  "select $unknown:foo_col from test",
			err:    true,
		},
		{
			desc:   "invalid, single arg, unknown alias, extra spaces",
			driver: DriverPostgres,
			query:  " select  $unknown:foo_col  from    test ",
			err:    true,
		},
		{
			desc:   "invalid, multiple args, unknown alias",
			driver: DriverPostgres,
			query:  "select $foo:foo_col, $unknown:bar_col from test",
			err:    true,
		},
		{
			desc:   "invalid, multiple args, unknown alias, extra spaces",
			driver: DriverPostgres,
			query:  "select $foo:foo_col ,    $unknown:bar_col   from  test ",
			err:    true,
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			cfg, err := testconfig.MakeTestConfigLite(test.driver)
			testcheck.FatalIf(t, err)
			qp := newQueryParser(cfg.DB.Driver, p, nil, cfg.OutMessages, test.autoMapOutMessages)
			q, stringers, err := qp.parseOutMessageArgs(test.query)
			testcheck.FatalIfUnexpected(t, err, test.err)
			if test.err {
				return
			}
			if q != test.expectedQuery {
				t.Fatalf("expected query to be %q but it was %q", test.expectedQuery, q)
			}
			stringerKeys := map[string]struct{}{}
			for k := range stringers {
				stringerKeys[k] = struct{}{}
			}
			if !eq.StringSets(stringerKeys, test.expectedStringerKeys) {
				t.Fatalf("expected %v stringer keys but got %v", test.expectedStringerKeys, stringerKeys)
			}
		})
	}
}
