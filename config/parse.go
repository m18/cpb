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
	fileName, flagConfig, err := p.parseFlags()
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}
	fileConfig, err := p.parseFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}
	flagConfig.merge(fileConfig)
	return p.from(flagConfig)
}

func (p *parser) parseFlags() (fileName string, flagsConfig *rawConfig, err error) {
	flagsConfig = newRawConfig()
	defaultSet := flag.NewFlagSet("config", flag.ContinueOnError)
	defaultSet.StringVar(&fileName, flagFile, "", fmt.Sprintf("Name of a config file to use. If not provided, an optional %q is assumed", defaultConfigFileName))
	defaultSet.StringVar(&flagsConfig.Protoc, flagProtoc, "", fmt.Sprintf("Path to protoc. If not provided, %q is assumed", defaultProtoc))
	defaultSet.StringVar(&flagsConfig.DB.Driver, flagDriver, "", "Database driver name. Possible values: postgres")
	defaultSet.StringVar(&flagsConfig.DB.Host, flagHost, "", "Host name or IP address")
	defaultSet.IntVar(&flagsConfig.DB.Port, flagPort, 0, "Port number")
	defaultSet.StringVar(&flagsConfig.DB.Name, flagName, "", "Database name")
	defaultSet.StringVar(&flagsConfig.DB.UserName, flagUserName, "", "User name")
	defaultSet.StringVar(&flagsConfig.DB.Password, flaagPassword, "", "Password")
	if p.mute {
		defaultSet.SetOutput(io.Discard)
	}
	if err = defaultSet.Parse(p.args); err != nil {
		return "", nil, err // possible flag.ErrHelp // TODO: handle it in main
	}
	return fileName, flagsConfig, nil
}

func (p *parser) parseFile(fileName string) (*rawConfig, error) {
	var optional bool
	if fileName == "" {
		// check the default config file;
		// if it does not exists, it's OK
		fileName = defaultConfigFileName
		optional = true
	}
	dfs := p.makeFS(filepath.Dir(fileName))
	f, err := dfs.Open(filepath.Base(fileName))
	if err != nil {
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("could not open file %q: %w", fileName, err)
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return nil, fmt.Errorf("not a file: %q", fileName)
	}
	fileConfig, err := newRawConfig().from(f)
	if err != nil {
		return nil, err
	}
	return fileConfig, nil
}

func (p *parser) from(raw *rawConfig) (res *Config, err error) {
	if raw == nil {
		return nil, nil
	}
	res = &Config{}
	res.Protoc = raw.Protoc
	res.DB = raw.DB
	if res.InMessages, err = p.in.parse(raw.Messages.In); err != nil {
		return nil, err
	}
	if res.OutMessages, err = p.out.parse(raw.Messages.Out); err != nil {
		return nil, err
	}
	return res, nil
}
