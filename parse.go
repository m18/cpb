package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TODO: package parse or data

var inParamReplacers = map[string]func() func() string{
	"postgres": func() func() string {
		counter := 0
		return func() string {
			counter = counter + 1
			return fmt.Sprintf("$%d", counter)
		}
	},
}

type queryParser struct {
	protos          *protos
	inMessages      map[string]*InMessage
	inParamReplacer func() func() string
}

func newQueryParser(driver string, inMessages map[string]*InMessage, p *protos) *queryParser {
	return &queryParser{
		protos:          p,
		inMessages:      inMessages,
		inParamReplacer: inParamReplacers[driver], // driver has already been validated
	}
}

func (p *queryParser) Parse(q string) (string, [][]byte, map[string]func([]byte) (string, error), error) {
	var inMessageArgs [][]byte
	var prettyPrinters map[string]func([]byte) (string, error)
	var err error

	if q, inMessageArgs, err = p.parseInMessageArgs(q); err != nil {
		return "", nil, nil, err
	}

	if q, prettyPrinters, err = p.parseOutMessageArgs(q); err != nil {
		return "", nil, nil, err
	}

	return q, inMessageArgs, prettyPrinters, nil
}

func (p *queryParser) parseOutMessageArgs(q string) (string, map[string]func([]byte) (string, error), error) {

	return q, nil, nil
}

func (p *queryParser) parseInMessageArgs(q string) (string, [][]byte, error) {
	if !strings.Contains(q, "$") {
		// the query doesn't have anything to replace in the `where` clase,
		// just return as is
		return q, nil, nil
	}
	// TODO: unit-test - "string", number, num.ber, false, true, OR none, e.g., ()
	var queryrx = regexp.MustCompile(`\$(?P<alias>\w+)\((?P<args>((\s*("\w+"|\d+(.\d+)?|true|false)\s*,)*(\s*("\w+"|\d+(.\d+)?|true|false)\s*))|)\)`)
	var argsrx = regexp.MustCompile(`"\w+"|\d+(.\d+)?|true|false`)
	// fmt.Println(fns.FindAllString(q, -1))

	groups, ok := findAllGroups(queryrx, q)
	if !ok {
		return "", nil, fmt.Errorf("malformed query: %q", q)
	}

	queryArgs := make([][]byte, 0, len(groups))
	for _, group := range groups {
		alias := group["alias"]
		inMessage, ok := p.inMessages[alias]
		if !ok {
			return "", nil, fmt.Errorf("unknown alias in query: %q", alias)
		}

		args, _ := findAllMatches(argsrx, group["args"]) // ignoring ok as it's been verified by queryrx already, treat empty () as ok too
		jsonMessage, err := inMessage.JSON(args)
		if err != nil {
			return "", nil, err
		}
		queryArg, err := p.protos.protoBytes(inMessage.Name, jsonMessage)
		if err != nil {
			return "", nil, err
		}

		queryArgs = append(queryArgs, queryArg)
	}

	counter := 0
	query := queryrx.ReplaceAllStringFunc(q, func(b string) string {
		counter = counter + 1
		return fmt.Sprintf("$%d", counter)
	})

	return query, queryArgs, nil
}

// TODO: generic, extract into regexp.go + unit tests
// returns a slice of maps where each map is a groupName:groupValue set
func findAllGroups(r *regexp.Regexp, s string) ([]map[string]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	groupNames := r.SubexpNames()[1:]
	var res []map[string]string
	for _, matchGroups := range r.FindAllStringSubmatch(s, -1) {
		m := make(map[string]string, len(groupNames))
		for i, v := range matchGroups[1:] {
			groupName := groupNames[i]
			if len(groupName) == 0 { // unnamed group
				groupName = strconv.Itoa(i)
			}
			m[groupName] = v
		}
		res = append(res, m)
	}
	return res, true
}

// TODO: generic, extract into regexp.go + unit tests
func findGroups(r *regexp.Regexp, s string) (map[string]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	groupNames := r.SubexpNames()[1:]
	res := make(map[string]string, len(groupNames))
	for i, v := range r.FindStringSubmatch(s)[1:] {
		groupName := groupNames[i]
		if len(groupName) == 0 { // unnamed group
			groupName = strconv.Itoa(i)
		}
		res[groupName] = v
	}
	return res, true
}

// TODO: generic, extract into regexp.go + unit tests
func findAllMatches(r *regexp.Regexp, s string) ([]string, bool) {
	if !r.MatchString(s) {
		return nil, false
	}
	return r.FindAllString(s, -1), true
}

// TODO: generic, extract into regexp.go + unit tests
func findMatch(r *regexp.Regexp, s string) (string, bool) {
	if !r.MatchString(s) {
		return "", false
	}
	return r.FindString(s), true
}

// TODO: generic, extract into regexp.go + unit tests
func replaceAllGroupsFunc(r *regexp.Regexp, s string, replace func(map[string]string) string) string {
	if !r.MatchString(s) {
		return s
	}
	res := ""
	lastIdx := 0
	groupNames := r.SubexpNames()[1:]

	for _, idxs := range r.FindAllSubmatchIndex([]byte(s), -1) {
		groups := map[string]string{}
		for i := 2; i < len(idxs); i += 2 { // skip the overall match
			groupIdx := (i - 2) / 2
			groupName := groupNames[groupIdx]
			if len(groupName) == 0 { // unnamed group
				groupName = strconv.Itoa(groupIdx + 1)
			}
			if idxs[i] == -1 || idxs[i+1] == -1 { // optional group is missing
				groups[groupName] = ""
			} else {
				groups[groupName] = s[idxs[i]:idxs[i+1]]
			}
		}
		res += s[lastIdx:idxs[0]] + replace(groups)
		lastIdx = idxs[1]
	}

	return res + s[lastIdx:]
}
