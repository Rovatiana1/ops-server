package ldap

type Config struct {
	Host       string `json:"host"`
	Port       int    `json:"port"`
	BaseDN     string `json:"baseDN"`
	BindDN     string `json:"bindDN"`
	Password   string `json:"password"`
	UserFilter string `json:"userFilter"`

	Attributes struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"attributes"`

	TLS     bool `json:"tls"`
	Timeout int  `json:"timeout"`
}
