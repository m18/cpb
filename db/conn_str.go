package db

import (
	"fmt"
	"net/url"

	"github.com/m18/cpb/config"
)

var connStrGens = map[string]func(*config.DBConfig) (string, error){
	// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS
	"postgres": func(c *config.DBConfig) (string, error) {
		if c.Port == 0 {
			c.Port = 5432
		}
		s := fmt.Sprintf("postgres://%s@%s:%d/%s",
			url.UserPassword(c.UserName, c.Password),
			url.QueryEscape(c.Host),
			c.Port,
			url.QueryEscape(c.Name),
		)
		u, err := url.Parse(s)
		if err != nil {
			return "", err
		}
		q := u.Query()
		for k, v := range c.Params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		return u.String(), nil
	},
}
