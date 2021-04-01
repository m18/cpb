package db

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/m18/cpb/config"
	"github.com/m18/cpb/protos"
	"github.com/m18/cpb/rx"
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
	inqueryrx, inargrx, outqueryrx *regexp.Regexp
	normalizeInMessageArgs         func([]string) []string
}

func newQueryParser(driver string, p *protos.Protos, inMessages map[string]*config.InMessage, outMessages map[string]*config.OutMessage) *queryParser {
	inargnormrx := regexp.MustCompile(`'((\\'|[^'])*)'`)
	normalizer := func(args []string) []string {
		for i, arg := range args {
			if inargnormrx.MatchString(arg) {
				arg = inargnormrx.ReplaceAllString(arg, "\"$1\"")
				args[i] = strings.ReplaceAll(arg, "\\'", "'")
			}
		}
		return args
	}

	return &queryParser{
		protos:                 p,
		inMessages:             inMessages,
		inParamReplacer:        inParamReplacers[driver], // driver has already been validated
		outMessages:            outMessages,
		inqueryrx:              regexp.MustCompile(`\$(?P<alias>\w+)\((?P<args>((\s*('(\\'|[^'])*'|\d+(.\d+)?|true|false)\s*,)*(\s*('(\\'|[^'])*'|\d+(.\d+)?|true|false)\s*))|)\)`),
		inargrx:                regexp.MustCompile(`'(\\'|[^'])*'|\d+(.\d+)?|true|false`),
		outqueryrx:             regexp.MustCompile(`\$(?P<alias>\w+):(?P<col>\w+|"(\w+\s*)+")(?P<full_col_alias>(\s+[aA][sS])?\s+(?P<col_alias>\w+|"(\w+\s*)+")[\s,$])?`),
		normalizeInMessageArgs: normalizer,
	}
}

// func (p *queryParser) Parse(q string) (string, [][]byte, map[string]func([]byte) (string, error), error) {
// 	var inMessageArgs [][]byte
// 	var prettyPrinters map[string]func([]byte) (string, error)
// 	var err error

// 	if q, inMessageArgs, err = p.parseInMessageArgs(q); err != nil {
// 		return "", nil, nil, err
// 	}

// 	if q, prettyPrinters, err = p.parseOutMessageArgs(q); err != nil {
// 		return "", nil, nil, err
// 	}

// 	return q, inMessageArgs, prettyPrinters, nil
// }

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
