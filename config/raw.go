package config

import (
	"encoding/json"
	"io"
)

type rawConfig struct {
	DB       *DBConfig       `json:"db"`
	Messages *messagesConfig `json:"messages"`
}

type messagesConfig struct {
	In  map[string]*inMessageConfig  `json:"in"`
	Out map[string]*outMessageConfig `json:"out"`
}

type inMessageConfig struct {
	Name     string      `json:"name"`     // TODO: change to protoreflect.FullName
	Template interface{} `json:"template"` // for JSON objects, map[string]interface{} behind the interface{} type
}

type outMessageConfig struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

func (c *rawConfig) from(r io.Reader) error {
	data, err := io.ReadAll(r)
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
