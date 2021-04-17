package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"text/template"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	defaultConfigFileName = "config.json"
	defaultProtoc         = "protoc"
	flagProtoc            = "c"
	flagProtoDir          = "b"
	flagDriver            = "d"
	flagHost              = "s"
	flagPort              = "p"
	flagName              = "n"
	flagUserName          = "u"
	flagPassword          = "w"

	FlagFile = "f"
)

// Config is application configuration.
type Config struct {
	Proto *Proto
	DB    *DBConfig

	InMessages  map[string]*InMessage
	OutMessages map[string]*OutMessage
}

// Proto encapsulates protobuf-specific configuration.
type Proto struct {
	C   string `json:"c"`
	Dir string `json:"dir"`
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

	Template *template.Template
	Props    map[string]struct{} // all dotProps defined in template
}

// New initializes and returns a new Config
func New(args []string, makeFS func(string) fs.FS) (*Config, error) {
	p := newParser(args, makeFS, false)
	res, err := p.parse()
	if err != nil {
		return nil, fmt.Errorf("could not create config: %w", err)
	}
	// TODO: wrap sentinel validation error
	if err = res.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return res, nil
}

// JSON template.
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

func (c *Config) validate() error {
	if c.Proto.C == "" {
		c.Proto.C = defaultProtoc
	}
	if c.DB.Driver == "" {
		return errors.New("driver is not specified")
	}
	if c.DB.Host == "" {
		return errors.New("host is not specified")
	}
	if c.DB.Port == 0 {
		return errors.New("port is not specified")
	}
	if c.DB.Name == "" {
		return errors.New("database name is not specified")
	}
	if c.DB.UserName == "" {
		return errors.New("user name is not specified")
	}
	return nil
}
