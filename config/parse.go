package config

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
)

type parser struct {
	args   []string
	makeFS func(string) fs.FS
	in     *inMessageParser
	out    *outMessageParser
	mute   bool
}

func newParser(args []string, makeFS func(string) fs.FS, mute bool) *parser {
	return &parser{
		args:   args,
		makeFS: makeFS,
		in:     newInMessageParser(),
		out:    newOutMessageParser(),
		mute:   mute,
	}
}

func (p *parser) parse() (*Config, error) {
	filePath, clConfig, isSet, err := p.parseCLArgs()
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	fileConfig, err := p.parseFile(filePath, isSet(FlagFile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	fileConfig.merge(clConfig, isSet)
	return p.from(fileConfig)
}

func (p *parser) parseCLArgs() (filePath string, flagsConfig *rawConfig, isSet func(string) bool, err error) {
	flagsConfig = newRawConfig()
	defaultSet := flag.NewFlagSet("config", flag.ContinueOnError)
	defaultSet.StringVar(&filePath, FlagFile, "", fmt.Sprintf("Path to a config file to use. If not provided, an optional %q is assumed", defaultConfigFileName))
	defaultSet.StringVar(&flagsConfig.Proto.C, flagProtoc, "", fmt.Sprintf("Path to protoc. If not provided, %q is assumed", defaultProtoc))
	defaultSet.StringVar(&flagsConfig.Proto.Dir, flagProtoDir, "", "Protobuf source root directory.")
	defaultSet.StringVar(&flagsConfig.DB.Driver, flagDriver, "", "Database driver name. Possible values: postgres")
	defaultSet.StringVar(&flagsConfig.DB.Host, flagHost, "", "Host name or IP address")
	defaultSet.IntVar(&flagsConfig.DB.Port, flagPort, 0, "Port number")
	defaultSet.StringVar(&flagsConfig.DB.Name, flagName, "", "Database name")
	defaultSet.StringVar(&flagsConfig.DB.UserName, flagUserName, "", "User name")
	defaultSet.StringVar(&flagsConfig.DB.Password, flagPassword, "", "Password")
	noAutoMap := defaultSet.Bool(flagNoAutoMap, false, "Do not auto-decode values in columns whose names match message aliases")
	if p.mute {
		defaultSet.SetOutput(io.Discard)
	}
	if err = defaultSet.Parse(p.args); err != nil {
		return "", nil, nil, err // possible flag.ErrHelp // TODO: handle it in main
	}
	flagsConfig.DB.Query = defaultSet.Arg(0)
	flagsConfig.Messages.AutoMap = !*noAutoMap

	m := map[string]struct{}{}
	defaultSet.Visit(func(f *flag.Flag) {
		m[f.Name] = struct{}{}
	})
	isSet = func(name string) bool {
		_, ok := m[name]
		return ok
	}

	return filePath, flagsConfig, isSet, nil
}

func (p *parser) parseFile(filePath string, isSet bool) (*rawConfig, error) {
	res := newRawConfig()
	if !isSet {
		filePath = defaultConfigFileName
	}
	fileName := filepath.Base(filePath)
	fsys := p.makeFS(filepath.Dir(filePath))
	if _, err := fs.Stat(fsys, fileName); err != nil {
		if !isSet {
			// default config file does not exists
			// it's OK
			return res, nil
		}
		return nil, fmt.Errorf("could not open file %q: %w", filePath, err)
	}
	bytes, err := fs.ReadFile(fsys, fileName)
	if err != nil {
		return nil, fmt.Errorf("could not read file %q: %w", filePath, err)
	}
	if err := res.from(bytes); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *parser) from(raw *rawConfig) (res *Config, err error) {
	if raw == nil {
		return nil, nil
	}
	res = &Config{}
	res.Proto = raw.Proto
	res.DB = raw.DB
	if res.InMessages, err = p.in.parse(raw.Messages.In); err != nil {
		return nil, err
	}
	if res.OutMessages, err = p.out.parse(raw.Messages.Out); err != nil {
		return nil, err
	}
	res.AutoMapOutMessages = raw.Messages.AutoMap
	return res, nil
}
