package ldap

import (
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type Client struct{}

func (c *Client) Connect(cfg *Config) (*ldap.Conn, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	var conn *ldap.Conn
	var err error

	if cfg.TLS {
		conn, err = ldap.DialTLS("tcp", addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	} else {
		conn, err = ldap.Dial("tcp", addr)
	}

	if err != nil {
		return nil, err
	}

	// bind
	if err := conn.Bind(cfg.BindDN, cfg.Password); err != nil {
		return nil, err
	}

	return conn, nil
}
