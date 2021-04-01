package db

import (
	"fmt"
	"testing"

	"github.com/m18/cpb/check"
	"github.com/m18/cpb/internal/testconfig"
	"github.com/m18/cpb/internal/testprotos"
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
	qp := newQueryParser(DriverPostgres, nil, nil, nil)
	for _, test := range tests {
		test := test
		t.Run(fmt.Sprint(test.args), func(t *testing.T) {
			t.Parallel()
			args := qp.normalizeInMessageArgs(test.args)
			if !check.StringSlicesAreEqual(args, test.expectedArgs) {
				t.Fatalf("expected %v but got %v", test.expectedArgs, args)
			}
		})
	}

}

func TestQueryParserParseInMessageArgs(t *testing.T) {
	p, err := testprotos.MakeProtosLite()
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		desc             string
		driver           string
		query            string
		expectedQuery    string
		expectedArgCount int
		err              bool
	}{
		{
			desc:             "valid, no params",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = 100",
			expectedQuery:    "select * from test where foo_col = 100",
			expectedArgCount: 0,
		},
		{
			desc:             "valid, malformed param",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1",
			expectedQuery:    "select * from test where foo_col = $foo(1",
			expectedArgCount: 0,
		},
		{
			desc:             "valid, single param",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single param, type coersion - '1' -> 1",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo('1', 'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single param, extra spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col   =  $foo(  1 , 'one' , true )  ",
			expectedQuery:    "select * from test where foo_col   =  $1  ",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single param, no spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col=$foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col=$1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single param, no left space",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col= $foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col= $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, single param, no right space",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col =$foo(1,'one',true)",
			expectedQuery:    "select * from test where foo_col =$1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, multiple params",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'one', true) and bar_col = $bar(2, 'two')",
			expectedQuery:    "select * from test where foo_col = $1 and bar_col = $2",
			expectedArgCount: 2,
		},
		{
			desc:             "valid, multiple params, mixed spaces",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col=$foo(   1   ,  'one'   , true )  and   bar_col =$bar( 2, 'two' ) ",
			expectedQuery:    "select * from test where foo_col=$1  and   bar_col =$2 ",
			expectedArgCount: 2,
		},
		{
			desc:             "valid, single param, escaped quote",
			driver:           DriverPostgres,
			query:            "select * from test where foo_col = $foo(1, 'o\\'one', true)",
			expectedQuery:    "select * from test where foo_col = $1",
			expectedArgCount: 1,
		},
		{
			desc:             "valid, multiple params, escaped quotes",
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
			desc:   "invalid, single param, wrong arg type (string instead of int32)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo('not a number', 'one', false)",
			err:    true,
		},
		{
			desc:   "invalid, single param, wrong arg type (string instead of bool)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one', 'false')",
			err:    true,
		},
		{
			desc:   "invalid, single param, wrong arg type (float instead of int32)",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1.1, 'one', false)",
			err:    true,
		},
		{
			desc:   "invalid, single param, wrong arg count",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1)",
			err:    true,
		},
		{
			desc:   "invalid, multiple params, wrong arg type",
			driver: DriverPostgres,
			query:  "select * from test where foo_col = $foo(1, 'one', false) and bar_col = $bar('not a number', 'two')",
			err:    true,
		},
		{
			desc:   "invalid, multiple params, wrong arg count",
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
			if err != nil {
				t.Fatal(err)
			}
			qp := newQueryParser(cfg.DB.Driver, p, cfg.InMessages, nil)
			q, args, err := qp.parseInMessageArgs(test.query)
			if err == nil == test.err {
				t.Fatalf("expected %t but did not get it: %v", test.err, err)
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
