package config

import (
	"encoding/json"

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

func newRawConfig() *rawConfig {
	return &rawConfig{
		DB:       &DBConfig{},
		Messages: &messagesConfig{},
	}
}

func (c *rawConfig) from(b []byte) (*rawConfig, error) {
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// merge merges secondary into c, with c values taking precedence over secondary values.
//
// c is never nil.
func (c *rawConfig) merge(secondary *rawConfig) {
	if secondary == nil {
		return
	}
	c.mergeString(&c.Protoc, secondary.Protoc)
	c.mergeString(&c.DB.Driver, secondary.DB.Driver)
	c.mergeString(&c.DB.Host, secondary.DB.Host)
	c.mergeInt(&c.DB.Port, secondary.DB.Port)
	c.mergeString(&c.DB.Name, secondary.DB.Name)
	c.mergeString(&c.DB.UserName, secondary.DB.UserName)
	c.mergeString(&c.DB.Password, secondary.DB.Password)

	// TODO: handle DB.Params & Messages item by item
	if len(c.DB.Params) == 0 {
		c.DB.Params = secondary.DB.Params
	}
	if len(c.Messages.In) == 0 {
		c.Messages.In = secondary.Messages.In
	}
	if len(c.Messages.Out) == 0 {
		c.Messages.Out = secondary.Messages.Out
	}

	// TODO: merge ENV vars
	// 		 custom field tags `env:"blah"`
	//		 os.LookupEnv("")
}

func (c *rawConfig) mergeString(target *string, v string) {
	if v == "" || *target != "" {
		return
	}
	*target = v
}

func (c *rawConfig) mergeInt(target *int, v int) {
	if v == 0 || *target != 0 {
		return
	}
	*target = v
}
