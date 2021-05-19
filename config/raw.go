package config

import (
	"encoding/json"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type rawConfig struct {
	Proto    *Proto          `json:"proto"`
	DB       *DBConfig       `json:"db"`
	Messages *messagesConfig `json:"messages"`
}

type messagesConfig struct {
	In      map[string]*inMessageConfig  `json:"in"`
	Out     map[string]*outMessageConfig `json:"out"`
	AutoMap bool                         `json:"autoMap"`
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
		Proto: &Proto{},
		DB:    &DBConfig{},
		Messages: &messagesConfig{
			AutoMap: true,
		},
	}
}

func (c *rawConfig) from(b []byte) error {
	if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

func (c *rawConfig) merge(override *rawConfig, isSet func(string) bool) {
	mergeString(&c.Proto.C, override.Proto.C, isSet(flagProtoc))
	mergeString(&c.Proto.Dir, override.Proto.Dir, isSet(flagProtoDir))
	mergeString(&c.DB.Driver, override.DB.Driver, isSet(flagDriver))
	mergeString(&c.DB.Host, override.DB.Host, isSet(flagHost))
	mergeInt(&c.DB.Port, override.DB.Port, isSet(flagPort))
	mergeString(&c.DB.Name, override.DB.Name, isSet(flagName))
	mergeString(&c.DB.UserName, override.DB.UserName, isSet(flagUserName))
	mergeString(&c.DB.Password, override.DB.Password, isSet(flagPassword))
	mergeBool(&c.Messages.AutoMap, override.Messages.AutoMap, isSet(flagNoAutoMap))

	// // TODO: handle DB.Params & Messages item by item
	// if len(c.DB.Params) == 0 {
	// 	c.DB.Params = secondary.DB.Params
	// }
	// if len(c.Messages.In) == 0 {
	// 	c.Messages.In = secondary.Messages.In
	// }
	// if len(c.Messages.Out) == 0 {
	// 	c.Messages.Out = secondary.Messages.Out
	// }

	// // TODO: merge ENV vars
	// // 		 custom field tags `env:"blah"`
	// //		 os.LookupEnv("")
}

func mergeString(target *string, v string, isSet bool) {
	if isSet {
		*target = v
	}
}

func mergeInt(target *int, v int, isSet bool) {
	if isSet {
		*target = v
	}
}

func mergeBool(target *bool, v bool, isSet bool) {
	if isSet {
		*target = v
	}
}
