package config

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/m18/cpb/internal/tmpl"
	"github.com/m18/cpb/rx"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	defaultConfigFileName = "config.json"
)

// Config is application configuration.
type Config struct {
	Protoc string
	DB     *DBConfig

	InMessages  map[string]*InMessage
	OutMessages map[string]*OutMessage
}

// InMessage is configuration for "in" messages, that is, messages going to the database.
type InMessage struct {
	Alias string
	Name  protoreflect.FullName

	template *template.Template
	params   []string
}

// OutMessage is configuration for "out" messages, that is, messages coming from the database.
type OutMessage struct {
	Alias string
	Name  protoreflect.FullName

	template *template.Template
	props    map[string]struct{} // all dotProps defined in template
}

// JSON representation of InMessage.
// TODO: Rename to String
func (m *InMessage) JSON(args []string) (string, error) {
	if len(args) != len(m.params) {
		return "", fmt.Errorf("wrong argument count for alias %q: %v", m.Alias, args)
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

// DBConfig encapsulates database configuration.
type DBConfig struct {
	Driver   string            `json:"driver"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Name     string            `json:"name"`
	UserName string            `json:"userName"`
	Password string            `json:"password"`
	Params   map[string]string `json:"params"`
	Query    string            `json:"query"`
}

// New initializes and returns a new Config
func New() (*Config, error) {
	argsFile := []string{}
	argsRest := []string{}
	for i := 1; i < len(os.Args); i++ {
		f := os.Args[i]
		if strings.HasPrefix(f, "-f") {
			argsFile = append(argsFile, f)
			if f == "-f" && i < len(os.Args)-1 && !strings.HasPrefix(os.Args[i+1], "-") {
				argsFile = append(argsFile, os.Args[i+1])
				i++
			}
			continue
		}
		argsRest = append(argsRest, f)
	}

	var fileName string
	fileSet := flag.NewFlagSet("file", flag.ExitOnError)
	fileSet.StringVar(&fileName, "f", defaultConfigFileName, "Name of a config file to use (default: \"config.json\").")
	// TODO: Go 1.16
	// fileSet.SetOutput(ioutil.Discard)
	if err := fileSet.Parse(argsFile); err != nil {
		return nil, err
	}

	dfs := os.DirFS(filepath.Dir(fileName))

	var raw rawConfig
	if err := raw.from(dfs, filepath.Base(fileName)); err != nil {
		return nil, err
	}

	set := flag.NewFlagSet("config", flag.ExitOnError)

	set.StringVar(&raw.Protoc, "c", defaultValue(raw.Protoc, "protoc"), "Path to protoc.")
	set.StringVar(&raw.DB.Driver, "d", raw.DB.Driver, "Database driver name. Possible values: postgres.")
	set.StringVar(&raw.DB.Host, "h", raw.DB.Host, "Host name or IP address.")
	set.IntVar(&raw.DB.Port, "p", raw.DB.Port, "Port number.")
	set.StringVar(&raw.DB.UserName, "u", raw.DB.UserName, "User name.")
	set.StringVar(&raw.DB.Password, "w", raw.DB.Password, "Password.")
	// TODO: custom param
	// flag.StringVar(&db.Params, "x", "", "Extra parameters in the \"param1=value1,param2=value2...\"format, e.g., \"sslmode=disable,connect_timeout=10\".")
	set.StringVar(&raw.DB.Query, "q", raw.DB.Query, "Query/command to execute.")

	set.Parse(argsRest)
	// TODO: use env vars on top of flags (custom field tags `env:"blah"`)
	//os.LookupEnv("")

	var res Config
	if err := res.from(&raw); err != nil {
		return nil, err
	}

	if err := res.validate(); err != nil {
		return nil, err
	}
	return &res, nil
}

func defaultValue(first, second string) string {
	if len(first) > 0 {
		return first
	}
	return second
}

func (c *Config) from(raw *rawConfig) error {
	res := Config{}
	res.Protoc = raw.Protoc
	res.DB = raw.DB
	if err := res.initInMessages(raw.Messages.In); err != nil {
		return err
	}
	if err := res.initOutMessages(raw.Messages.Out); err != nil {
		return err
	}
	*c = res
	return nil
}

func (c *Config) validate() error {
	// TODO: implement
	return nil
}

// TODO: unit-test
func (c *Config) initInMessages(m map[string]*inMessageConfig) (err error) {
	var aliasrx = regexp.MustCompile(`^\s*(?P<alias>\w+)\s*\((?P<params>((\s*\w+\s*,)*\s*\w+\s*)|)\)$`)
	var paramsrx = regexp.MustCompile(`\w+`)
	var tplrx = regexp.MustCompile(`:\s*"\$(?P<varname>\w+)"`)

	res := make(map[string]*InMessage, len(m))

	for k, v := range m {
		// parse alias
		groups, ok := rx.FindGroups(aliasrx, k)
		if !ok {
			return fmt.Errorf("invalid alias definition: %q", k)
		}

		im := &InMessage{Alias: groups["alias"], Name: v.Name}

		// parse alias params
		pg := groups["params"]
		im.params, _ = rx.FindAllMatches(paramsrx, pg) // ignoring ok as it's been verified by aliasrx already, treat empty () as ok too
		paramLookup := map[string]struct{}{}
		for _, p := range im.params {
			if _, ok := paramLookup[p]; ok {
				return fmt.Errorf("duplicate parameter name for alias %q: %q", im.Alias, p)
			}
			paramLookup[p] = struct{}{}
		}

		// parse tpl
		tplbytes, err := json.Marshal(v.Template) // marshal interface{} (map[string]interface{} for JSON objects) back to string
		if err != nil {
			return err
		}
		tpl := string(tplbytes)
		allGroups, _ := rx.FindAllGroups(tplrx, tpl) // ok for there to be no matches
		for _, groups := range allGroups {
			vn := groups["varname"]
			if _, ok := paramLookup[vn]; !ok {
				return fmt.Errorf("unknown variable name for alias %q: %q", im.Alias, vn)
			}
		}
		tpl = tplrx.ReplaceAllString(tpl, ":{{.$varname}}")
		if im.template, err = template.New(im.Alias).Parse(tpl); err != nil {
			return fmt.Errorf("invalid message template: %q", v.Template)
		}

		if _, ok := res[im.Alias]; ok {
			return fmt.Errorf("duplicate alias: %q", im.Alias)
		}
		res[im.Alias] = im
	}

	c.InMessages = res
	return nil
}

// TODO: unit-test
func (c *Config) initOutMessages(m map[string]*outMessageConfig) (err error) {
	// TODO: add $ at the end - `^\w+$`
	var aliasrx = regexp.MustCompile(`^\w+`)
	var tplrx = regexp.MustCompile(`(?P<prefix>[^\\]|^)(?P<marker>\$)(?P<prop>(\w+\.)*\w+)`) // $ can be escaped with with \$ (\\$ in json)
	res := make(map[string]*OutMessage, len(m))
	for k, v := range m {
		// alias
		alias, ok := rx.FindMatch(aliasrx, k)
		if !ok {
			return fmt.Errorf("invalid alias definition: %q", k)
		}

		// template
		props := map[string]struct{}{}
		tpl := rx.ReplaceAllGroupsFunc(tplrx, v.Template, func(groups map[string]string) string {
			prop := groups["prop"]
			props[prop] = struct{}{}
			return groups["prefix"] + "{{." + tmpl.PropToTemplateParam(prop) + "}}"
		})
		tpl = strings.ReplaceAll(tpl, "\\$", "$") // unescape any `\$`s after rx-replace is done

		om := &OutMessage{Alias: alias, Name: v.Name, props: props}

		if om.template, err = template.New(om.Alias).Parse(tpl); err != nil {
			return fmt.Errorf("invalid message template: %q", v.Template)
		}

		// TODO: can be done earlier
		if _, ok := res[om.Alias]; ok {
			return fmt.Errorf("duplicate alias: %q", om.Alias)
		}
		res[om.Alias] = om
	}

	c.OutMessages = res
	return nil
}
