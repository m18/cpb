package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"text/template"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	inMessages = "in-messages.json"
)

type config struct {
	Driver  string
	ConnStr string
	Query   string

	InMessages map[string]*InMessage
}

type InMessage struct {
	Alias string
	Type  protoreflect.FullName

	template *template.Template
	params   []string
}

func (m *InMessage) JSON(args []string) (string, error) {
	if len(args) != len(m.params) {
		return "", fmt.Errorf("incorrect argument count for alias %q: %v", m.Alias, args)
	}
	data := map[string]string{}
	for i, paramName := range m.params {
		data[paramName] = args[i]
	}
	var buf bytes.Buffer
	if err := m.template.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type InMessageConfig struct {
	Msg string `json:"msg"`
	Tpl string `json:"tpl"`
}

func newConfig() (*config, error) {
	var res config
	flag.StringVar(&res.Driver, "d", "", "Database driver name. Possible values: postgres.")
	flag.StringVar(&res.ConnStr, "c", "", "Database connection string.")
	flag.StringVar(&res.Query, "q", "", "Query/command to execute.")
	flag.Parse()
	// TODO: use env vars on top of flags
	//os.LookupEnv("")
	if err := res.parseInMessageConfig(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *config) validate() error {
	// TODO: implement
	return nil
}

// TODO: speed-up app start-up by reading each config file in a separate goroutine

func (c *config) parseInMessageConfig() error {
	// TODO: use 1.16
	data, err := ioutil.ReadFile(inMessages)
	if err != nil {
		return err
	}
	var m map[string]*InMessageConfig
	if err = json.Unmarshal(data, &m); err != nil {
		return err
	}

	if c.InMessages, err = c.createInMessages(m); err != nil {
		return err
	}

	return nil
}

// TODO: unit-test
func (c *config) createInMessages(m map[string]*InMessageConfig) (res map[string]*InMessage, err error) {
	var aliasrx = regexp.MustCompile(`^\s*(?P<alias>\w+)\s*\((?P<params>((\s*\w+\s*,)*\s*\w+\s*)|)\)$`)
	var paramsrx = regexp.MustCompile(`\w+`)
	var tplrx = regexp.MustCompile(`\$(?P<varname>\w+)(?P<followedBy>[\s,\}$])`) // only match full var names, e.g., $varname but not $varname@@@
	res = make(map[string]*InMessage, len(m))
	for k, v := range m {
		// parse alias
		groups, ok := findGroups(aliasrx, k)
		if !ok {
			return nil, fmt.Errorf("invalid alias definition: %q (%s)", k, inMessages)
		}
		group := groups[0] // only 1 set of groups as per aliasrx (no + or * on groups)
		im := &InMessage{Alias: group["alias"], Type: protoreflect.FullName(v.Msg)}

		// parse alias params
		pg := group["params"]
		params, _ := findMatches(paramsrx, pg) // ignoring ok as it's been verified by aliasrx already, treat empty () as ok too
		im.params = params
		paramsLookup := map[string]struct{}{}
		for _, p := range params {
			if _, ok := paramsLookup[p]; ok {
				return nil, fmt.Errorf("duplicate parameter name for alias %q: %q (%s)", im.Alias, p, inMessages)
			}
			paramsLookup[p] = struct{}{}
		}

		// parse tpl
		tpl := v.Tpl
		groups, ok = findGroups(tplrx, tpl)
		for _, group = range groups {
			vn := group["varname"]
			if _, ok := paramsLookup[vn]; !ok {
				return nil, fmt.Errorf("unknown variable name for alias %q: %q (%s)", im.Alias, vn, inMessages)
			}
		}
		tpl = tplrx.ReplaceAllString(tpl, "{{.$varname}}$followedBy")
		if im.template, err = template.New(im.Alias).Parse(tpl); err != nil {
			return nil, fmt.Errorf("invalid message template: %q (%s)", v.Tpl, inMessages)
		}

		if _, ok := res[im.Alias]; ok {
			return nil, fmt.Errorf("duplicate alias: %q (%s)", im.Alias, inMessages)
		}
		res[im.Alias] = im
	}
	return res, nil
}

// return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
// 	p.UserName,
// 	p.Password,
// 	p.Host,
// 	port,
// 	p.DBName,
