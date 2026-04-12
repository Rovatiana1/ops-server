package ldap

import (
	"context"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type LdapService interface {
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

type ldapService struct {
	provider ConfigProvider
	client   *Client
}

func NewService(p ConfigProvider) LdapService {
	return &ldapService{
		provider: p,
		client:   &Client{},
	}
}

type User struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
}

func (s *ldapService) loadConfig(ctx context.Context) (*Config, error) {
	return s.provider.Get(ctx)
}

func (s *ldapService) connect(cfg *Config) (*ldap.Conn, error) {
	return s.client.Connect(cfg)
}

func (s *ldapService) bindUser(conn *ldap.Conn, dn, password string) error {
	if err := conn.Bind(dn, password); err != nil {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}

func (s *ldapService) mapUser(entry *ldap.Entry, cfg *Config) *User {
	return &User{
		Username:  entry.GetAttributeValue(cfg.Attributes.Username),
		Email:     entry.GetAttributeValue(cfg.Attributes.Email),
		FirstName: entry.GetAttributeValue(cfg.Attributes.FirstName),
		LastName:  entry.GetAttributeValue(cfg.Attributes.LastName),
	}
}

func (s *ldapService) searchUser(conn *ldap.Conn, cfg *Config, username string) (*ldap.Entry, error) {
	filter := fmt.Sprintf(cfg.UserFilter, username)

	req := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		filter,
		[]string{
			cfg.Attributes.Username,
			cfg.Attributes.Email,
			cfg.Attributes.FirstName,
			cfg.Attributes.LastName,
		},
		nil,
	)

	res, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("ldap search failed: %w", err)
	}

	if len(res.Entries) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return res.Entries[0], nil
}

func (s *ldapService) TestConnection(ctx context.Context, cfg *Config) error {
	conn, err := s.connect(cfg)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Test bind avec compte technique
	if cfg.BindDN != "" && cfg.Password != "" {
		if err := conn.Bind(cfg.BindDN, cfg.Password); err != nil {
			return fmt.Errorf("bind failed: %w", err)
		}
	}

	// Test search simple
	req := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		"(objectClass=*)",
		[]string{"dn"},
		nil,
	)

	_, err = conn.Search(req)
	if err != nil {
		return fmt.Errorf("search test failed: %w", err)
	}

	return nil
}

func (s *ldapService) Authenticate(ctx context.Context, username, password string) (*User, error) {
	cfg, err := s.loadConfig(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := s.connect(cfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	entry, err := s.searchUser(conn, cfg, username)
	if err != nil {
		return nil, err
	}

	if err := s.bindUser(conn, entry.DN, password); err != nil {
		return nil, err
	}

	return s.mapUser(entry, cfg), nil
}
