package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// TODO: -flag
	configFileName = "config.json"
)

type config struct {
	Driver  string
	ConnStr string
	Query   string

	InMessages  map[string]*InMessage
	OutMessages map[string]*OutMessage
}

type InMessage struct {
	Alias string
	Name  protoreflect.FullName

	template *template.Template
	params   []string
}

type OutMessage struct {
	Alias string
	Name  protoreflect.FullName

	template *template.Template
	props    map[string]struct{} // all dotProps defined in template
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

type rawConfig struct {
	Messages messagesConfig `json:"messages"`
}

type messagesConfig struct {
	In  map[string]*inMessageConfig  `json:"in"`
	Out map[string]*outMessageConfig `json:"out"`
}

type inMessageConfig struct {
	Name     string      `json:"name"`
	Template interface{} `json:"template"` // for JSON objects, map[string]interface{} behind the interface{} type
}

type outMessageConfig struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

func newConfig() (*config, error) {
	var res config
	if err := res.parseFile(); err != nil {
		return nil, err
	}
	flag.StringVar(&res.Driver, "d", "", "Database driver name. Possible values: postgres.")
	flag.StringVar(&res.ConnStr, "c", "", "Database connection string.")
	flag.StringVar(&res.Query, "q", "", "Query/command to execute.")
	flag.Parse()
	// TODO: use env vars on top of flags
	//os.LookupEnv("")
	if err := res.validate(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *config) validate() error {
	// TODO: implement
	return nil
}

// TODO: speed-up app start-up by reading each config file in a separate goroutine

func (c *config) parseFile() error {
	// TODO: use 1.16
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return err
	}
	var raw rawConfig
	if err = json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if c.InMessages, err = c.createInMessages(raw.Messages.In); err != nil {
		return err
	}
	if c.OutMessages, err = c.createOutMessages(raw.Messages.Out); err != nil {
		return err
	}
	return nil
}

// TODO: unit-test
func (c *config) createInMessages(m map[string]*inMessageConfig) (res map[string]*InMessage, err error) {
	var aliasrx = regexp.MustCompile(`^\s*(?P<alias>\w+)\s*\((?P<params>((\s*\w+\s*,)*\s*\w+\s*)|)\)$`)
	var paramsrx = regexp.MustCompile(`\w+`)
	var tplrx = regexp.MustCompile(`:\s*"\$(?P<varname>\w+)"`)
	// var tplrx = regexp.MustCompile(`\$(?P<varname>\w+)(?P<followedBy>[\s,\}$])`) // only match full var names, e.g., $varname but not $varname@@@, $ matches end of line in multiline mode
	res = make(map[string]*InMessage, len(m))
	for k, v := range m {
		// parse alias
		groups, ok := findGroups(aliasrx, k)
		if !ok {
			return nil, fmt.Errorf("invalid alias definition: %q (%s)", k, configFileName)
		}

		im := &InMessage{Alias: groups["alias"], Name: protoreflect.FullName(v.Name)}

		// parse alias params
		pg := groups["params"]
		params, _ := findAllMatches(paramsrx, pg) // ignoring ok as it's been verified by aliasrx already, treat empty () as ok too
		im.params = params
		paramLookup := map[string]struct{}{}
		for _, p := range params {
			if _, ok := paramLookup[p]; ok {
				return nil, fmt.Errorf("duplicate parameter name for alias %q: %q (%s)", im.Alias, p, configFileName)
			}
			paramLookup[p] = struct{}{}
		}

		// parse tpl
		tplbytes, err := json.Marshal(v.Template) // marshal interface{} (map[string]interface{} for JSON objects) back to string
		if err != nil {
			return nil, err
		}
		tpl := string(tplbytes)
		allGroups, _ := findAllGroups(tplrx, tpl) // ok for there to be no matches
		for _, groups := range allGroups {
			vn := groups["varname"]
			if _, ok := paramLookup[vn]; !ok {
				return nil, fmt.Errorf("unknown variable name for alias %q: %q (%s)", im.Alias, vn, configFileName)
			}
		}
		tpl = tplrx.ReplaceAllString(tpl, ":{{.$varname}}")
		// tpl = tplrx.ReplaceAllString(tpl, "{{.$varname}}$followedBy")
		if im.template, err = template.New(im.Alias).Parse(tpl); err != nil {
			return nil, fmt.Errorf("invalid message template: %q (%s)", v.Template, configFileName)
		}

		if _, ok := res[im.Alias]; ok {
			return nil, fmt.Errorf("duplicate alias: %q (%s)", im.Alias, configFileName)
		}
		res[im.Alias] = im
	}
	return res, nil
}

// TODO: unit-test
func (c *config) createOutMessages(m map[string]*outMessageConfig) (res map[string]*OutMessage, err error) {
	// TODO: add $ at the end - `^\w+$`
	var aliasrx = regexp.MustCompile(`^\w+`)
	var tplrx = regexp.MustCompile(`(?P<prefix>[^\\]|^)(?P<marker>\$)(?P<prop>(\w+\.)*\w+)`) // $ can be escaped with with \$ (\\$ in json)
	res = make(map[string]*OutMessage, len(m))
	for k, v := range m {
		alias, ok := findMatch(aliasrx, k)
		if !ok {
			return nil, fmt.Errorf("invalid alias definition: %q (%s)", k, configFileName)
		}
		props := map[string]struct{}{}
		tpl := replaceAllGroupsFunc(tplrx, v.Template, func(groups map[string]string) string {
			prop := groups["prop"]
			props[prop] = struct{}{}
			return groups["prefix"] + "{{." + propToTplParam(prop) + "}}"
		})
		tpl = strings.ReplaceAll(tpl, "\\$", "$") // unescape any `\$`s after rx-replace is done

		om := &OutMessage{Alias: alias, Name: protoreflect.FullName(v.Name), props: props}

		if om.template, err = template.New(om.Alias).Parse(tpl); err != nil {
			return nil, fmt.Errorf("invalid message template: %q (%s)", v.Template, configFileName)
		}

		// TODO: can be done earlier
		if _, ok := res[om.Alias]; ok {
			return nil, fmt.Errorf("duplicate alias: %q (%s)", om.Alias, configFileName)
		}
		res[om.Alias] = om
	}
	return res, nil
}

// return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
// 	p.UserName,
// 	p.Password,
// 	p.Host,
// 	port,
// 	p.DBName,
