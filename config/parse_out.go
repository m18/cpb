package config

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/m18/cpb/internal/tmpl"
	"github.com/m18/cpb/rx"
)

type outMessageParser struct {
	aliasrx, tplrx *regexp.Regexp
}

func newOutMessageParser() *outMessageParser {
	return &outMessageParser{
		aliasrx: regexp.MustCompile(`^\w+$`),
		tplrx:   regexp.MustCompile(`(?P<prefix>[^\\]|^)(?P<marker>\$)(?P<prop>(\w+\.)*\w+)`), // $ can be escaped with with \$ (\\$ in json)
	}
}

func (p *outMessageParser) parse(m map[string]*outMessageConfig) (map[string]*OutMessage, error) {
	res := make(map[string]*OutMessage, len(m))
	for alias, omc := range m {
		om, err := p.parseMessage(alias, omc)
		if err != nil {
			return nil, err
		}
		// in case of duplicate keys in JSON, json.Unmarshal uses the last one,
		// so there can't be any duplicates since aliasrx doesn't allow any non-word characters
		// (e.g., keys like "foo " - if that was allowed, a key like " foo" or "foo" would be a duplicate after parsing/trimming)
		res[om.Alias] = om
	}
	return res, nil
}

func (p *outMessageParser) parseMessage(rawAlias string, omc *outMessageConfig) (*OutMessage, error) {
	alias, err := p.parseAlias(rawAlias)
	if err != nil {
		return nil, err
	}
	tpl, props, err := p.parseTemplate(alias, omc.Template)
	if err != nil {
		return nil, err
	}
	return &OutMessage{
		Alias:    alias,
		Name:     omc.Name,
		Template: tpl,
		Props:    props,
	}, nil
}

func (p *outMessageParser) parseAlias(alias string) (string, error) {
	res, ok := rx.FindMatch(p.aliasrx, alias)
	if !ok {
		return "", fmt.Errorf("invalid alias definition: %q", alias)
	}
	return res, nil
}

func (p *outMessageParser) parseTemplate(alias, tpl string) (res *template.Template, props map[string]struct{}, err error) {
	props = map[string]struct{}{}
	s := rx.ReplaceAllGroupsFunc(p.tplrx, tpl, func(groups map[string]string) string {
		prop := groups["prop"]
		props[prop] = struct{}{}
		return groups["prefix"] + "{{." + tmpl.PropToTemplateParam(prop) + "}}"
	})
	s = strings.ReplaceAll(s, "\\$", "$") // unescape any `\$`s after rx-replace is done
	res, err = template.New(alias).Parse(s)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid message template: %q", tpl)
	}
	return res, props, nil
}
