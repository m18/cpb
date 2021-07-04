package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/protos"
	"github.com/m18/rx"
)

var inParamReplacers = map[string]func() func(string) string{
	DriverPostgres: func() func(string) string {
		counter := 0
		return func(string) string {
			counter = counter + 1
			return fmt.Sprintf("$%d", counter)
		}
	},
}

type queryParser struct {
	protos                         *protos.Protos
	inMessages                     map[string]*config.InMessage
	inParamReplacer                func() func(string) string
	outMessages                    map[string]*config.OutMessage
	autoMapOutMessages             bool
	inqueryrx, inargrx, outqueryrx *regexp.Regexp
	normalizeInMessageArgs         func([]string) []string
}

func newQueryParser(driver string, p *protos.Protos, inMessages map[string]*config.InMessage, outMessages map[string]*config.OutMessage, autoMapOutMessages bool) *queryParser {
	inargnormrx := regexp.MustCompile(`'((\\'|[^'])*)'`) // checks for the presense of \' anywhere between a pair of single quotes
	normalizer := func(args []string) []string {         // performs transformations like 'A string' -> "A string", 'O\'Reilly' -> "O'Reilly"
		for i, arg := range args {
			if inargnormrx.MatchString(arg) {
				arg = inargnormrx.ReplaceAllString(arg, "\"$1\"") // replace enclosing '' with ""
				args[i] = strings.ReplaceAll(arg, "\\'", "'")     // replace \' with '
			}
		}
		return args
	}

	return &queryParser{
		protos:             p,
		inMessages:         inMessages,
		inParamReplacer:    inParamReplacers[driver], // driver has already been validated
		outMessages:        outMessages,
		autoMapOutMessages: autoMapOutMessages,
		inqueryrx:          regexp.MustCompile(`\$(?P<alias>\w+)\((?P<args>((\s*('(\\'|[^'])*'|\d+(.\d+)?|true|false)\s*,)*(\s*('(\\'|[^'])*'|\d+(.\d+)?|true|false)\s*))|)\)`),
		inargrx:            regexp.MustCompile(`'(\\'|[^'])*'|\d+(.\d+)?|true|false`),
		// TODO: postrges uses "" for reserved-word or space-separated col names; handle [] for sql server (and mysql?)
		outqueryrx:             regexp.MustCompile(`\$(?P<alias>\w+):(?P<col>\w+|"(\w+\s*)+")(?P<full_col_alias>(\s+[aA][sS])?\s+(?P<col_alias>\w+|"(\w+\s*)+")[\s,$])?`),
		normalizeInMessageArgs: normalizer,
	}
}

func (p *queryParser) parse(q string) (string, [][]byte, map[string]func([]byte) (string, error), error) {
	var inMessageArgs [][]byte
	var outMessageStringers map[string]func([]byte) (string, error)
	var err error

	if q, inMessageArgs, err = p.parseInMessageArgs(q); err != nil {
		return "", nil, nil, err
	}

	if q, outMessageStringers, err = p.parseOutMessageArgs(q); err != nil {
		return "", nil, nil, err
	}

	return q, inMessageArgs, outMessageStringers, nil
}

func (p *queryParser) parseInMessageArgs(q string) (string, [][]byte, error) {
	groups, ok := rx.FindAllGroups(p.inqueryrx, q)
	if !ok {
		// ok to have no "in" messages
		return q, nil, nil
	}
	queryArgs := make([][]byte, 0, len(groups))
	for _, group := range groups {
		alias := group["alias"]
		inMessage, ok := p.inMessages[alias]
		if !ok {
			return "", nil, fmt.Errorf("unknown alias in query: %q", alias)
		}

		args, _ := rx.FindAllMatches(p.inargrx, group["args"]) // ignoring "ok" as it's been verified by p.inqueryrx already; treat empty () as ok too
		args = p.normalizeInMessageArgs(args)
		jsonMessage, err := inMessage.JSON(args)
		if err != nil {
			return "", nil, err
		}
		queryArg, err := p.protos.ProtoBytes(inMessage.Name, jsonMessage)
		if err != nil {
			return "", nil, err
		}

		queryArgs = append(queryArgs, queryArg)
	}

	query := p.inqueryrx.ReplaceAllStringFunc(q, p.inParamReplacer())

	return query, queryArgs, nil
}

func (p *queryParser) parseOutMessageArgs(q string) (string, map[string]func([]byte) (string, error), error) {
	var err error
	var stringers map[string]func([]byte) (string, error)
	if p.autoMapOutMessages {
		if stringers, err = p.makeAutoOutMessageStringers(); err != nil {
			return "", nil, err
		}
	} else {
		stringers = map[string]func([]byte) (string, error){}
	}

	q = rx.ReplaceAllGroupsFunc(p.outqueryrx, q, func(groups map[string]string) string {
		if err != nil {
			return ""
		}

		alias := groups["alias"]
		col := groups["col"]
		colAlias := groups["col_alias"]

		outMessage, ok := p.outMessages[alias]
		if !ok {
			err = fmt.Errorf("unknown alias in query: %q", alias)
			return ""
		}

		var key string
		if len(colAlias) > 0 && !strings.EqualFold(colAlias, "from") {
			key = colAlias
		} else {
			key = col
		}

		// mapping is by col name, not col order (plus auto-mapping can only be done by col name)
		// this might be an issue in case there are multiple cols with the same name because of aliasing with AS
		// select *, $p:dat - currently will pretty-print both "dat" cols
		if stringers[key], err = p.protos.StringerFor(outMessage); err != nil {
			return ""
		}

		return col + groups["full_col_alias"]
	})

	if err != nil {
		return "", nil, err
	}
	return q, stringers, nil
}

func (p *queryParser) makeAutoOutMessageStringers() (map[string]func([]byte) (string, error), error) {
	var err error
	stringers := map[string]func([]byte) (string, error){}
	for alias, outMessage := range p.outMessages {
		if stringers[alias], err = p.protos.StringerFor(outMessage); err != nil {
			return nil, err
		}
	}
	return stringers, err
}
