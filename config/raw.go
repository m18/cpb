package config

import (
	"encoding/json"
	"io"
	"io/fs"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type rawConfig struct {
	Protoc   string          `json:"protoc"`
	DB       *DBConfig       `json:"db"`
	Messages *messagesConfig `json:"messages"`
}

type messagesConfig struct {
	In  map[string]*inMessageConfig  `json:"in"`
	Out map[string]*outMessageConfig `json:"out"`
}

type inMessageConfig struct {
	Name     protoreflect.FullName `json:"name"`
	Template interface{}           `json:"template"` // for JSON objects, map[string]interface{} behind the interface{} type
}

type outMessageConfig struct {
	Name     protoreflect.FullName `json:"name"`
	Template string                `json:"template"`
}

func (c *rawConfig) from(dfs fs.FS, fileName string) error {
	f, err := dfs.Open(fileName)
	if err != nil {
		return err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, c); err != nil {
		return err
	}
	if c.DB == nil {
		c.DB = &DBConfig{}
	}
	if c.Messages == nil {
		c.Messages = &messagesConfig{}
	}
	return nil
}
