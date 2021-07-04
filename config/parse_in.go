package config

import (
	"encoding/json"
	"fmt"
	"regexp"
	"text/template"

	"github.com/m18/rx"
)

type inMessageParser struct {
	aliasrx, paramsrx, tplrx *regexp.Regexp
}

func newInMessageParser() *inMessageParser {
	return &inMessageParser{
		aliasrx:  regexp.MustCompile(`^\s*(?P<alias>\w+)\s*\((?P<params>((\s*\w+\s*,)*\s*\w+\s*)|)\)$`),
		paramsrx: regexp.MustCompile(`\w+`),
		tplrx:    regexp.MustCompile(`:\s*"\$(?P<varname>\w+)"`),
	}
}

func (p *inMessageParser) parse(m map[string]*inMessageConfig) (map[string]*InMessage, error) {
	res := make(map[string]*InMessage, len(m))
	for aliasWithParams, imc := range m {
		im, err := p.parseMessage(aliasWithParams, imc)
		if err != nil {
			return nil, err
		}
		// only duplicate aliases with different param definition (param number, param names, different whitespacing, etc.) can be detected,
		// exactly the same aliases+params will not be treated as duplicates with the last one taking precedence.
		// see outMessageParser.parse for more details
		if _, ok := res[im.Alias]; ok {
			return nil, fmt.Errorf("duplicate in message alias: %q", im.Alias)
		}
		res[im.Alias] = im
	}
	return res, nil
}

func (p *inMessageParser) parseMessage(aliasWithParams string, imc *inMessageConfig) (*InMessage, error) {
	alias, aliasParams, err := p.parseAlias(aliasWithParams)
	if err != nil {
		return nil, err
	}
	params, paramLookup, err := p.parseAliasParams(alias, aliasParams)
	if err != nil {
		return nil, err
	}
	tpl, err := p.parseTemplate(alias, imc.Template, paramLookup)
	if err != nil {
		return nil, err
	}
	return &InMessage{
		Alias:    alias,
		Name:     imc.Name,
		template: tpl,
		params:   params,
	}, nil
}

func (p *inMessageParser) parseAlias(aliasWithParams string) (alias, aliasParams string, err error) {
	groups, ok := rx.FindGroups(p.aliasrx, aliasWithParams)
	if !ok {
		return "", "", fmt.Errorf("invalid alias definition: %q", aliasWithParams)
	}
	return groups["alias"], groups["params"], nil
}

func (p *inMessageParser) parseAliasParams(alias, aliasParams string) (params []string, paramLookup map[string]struct{}, err error) {
	params, _ = rx.FindAllMatches(p.paramsrx, aliasParams) // ignoring ok as it's been verified by aliasrx already, treat empty () as ok too
	paramLookup = map[string]struct{}{}
	for _, p := range params {
		if _, ok := paramLookup[p]; ok {
			return nil, nil, fmt.Errorf("duplicate parameter name for alias %q: %q", alias, p)
		}
		paramLookup[p] = struct{}{}
	}
	return params, paramLookup, nil
}

func (p *inMessageParser) parseTemplate(alias string, tplm interface{}, paramLookup map[string]struct{}) (*template.Template, error) {
	tplbytes, err := json.Marshal(tplm) // marshal interface{} (map[string]interface{} for JSON objects) back to string
	if err != nil {
		return nil, err
	}
	tpl := string(tplbytes)
	allGroups, _ := rx.FindAllGroups(p.tplrx, tpl) // ok for there to be no matches
	for _, groups := range allGroups {
		vn := groups["varname"]
		if _, ok := paramLookup[vn]; !ok {
			return nil, fmt.Errorf("unknown variable name for alias %q: %q", alias, vn)
		}
	}
	tpl = p.tplrx.ReplaceAllString(tpl, ":{{.$varname}}")

	res := template.Must(template.New(alias).Parse(tpl))
	return res, nil
}
