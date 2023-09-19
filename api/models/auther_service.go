package models

type KeycloakToken struct {
	AccessToken       string `json:"access_token"`
	ExpiresIn         int    `json:"expires_in"`
	IdToken           string `json:"id_token"`
	RefreshExpires_in int    `json:"refresh_expires_in"`
	RefreshToken      string `json:"refresh_token"`
	SessionState      string `json:"session_state"`
	TokenType         string `json:"token_type"`
}

type KeycloakUser struct {
	Email         string               `json:"email"`
	EmailVerified bool                 `json:"emailVerified"`
	Enabled       bool                 `json:"enabled"`
	FirstName     string               `json:"firstName"`
	Id            string               `json:"id"`
	LastName      string               `json:"lastName"`
	Username      string               `json:"username"`
	Attributes    *map[string][]string `json:"attributes"`
	Groups        *map[string][]string `json:"groups"`
}

type IntrospectSuccess struct {
	Exp int `json:"exp,omitempty" xml:"exp,omitempty"`
	Nbf int `json:"nbf,omitempty" xml:"nbf,omitempty"`
	Iat int `json:"iat,omitempty" xml:"iat,omitempty"`
	// Aud      string `json:"aud,omitempty" xml:"aud,omitempty"`
	Active      bool                 `json:"active,omitempty" xml:"active,omitempty"`
	AuthTime    int                  `json:"auth_time,omitempty" xml:"auth_time,omitempty"`
	Jti         string               `json:"jti,omitempty" xml:"jti,omitempty"`
	Type        string               `json:"typ,omitempty" xml:"typ,omitempty"`
	Permissions []ResourcePermission `json:"permissions,omitempty"`
}

type ResourcePermission struct {
	RSID           string   `json:"rsid,omitempty"`
	ResourceID     string   `json:"resource_id,omitempty"`
	RSName         string   `json:"rsname,omitempty"`
	Scopes         []string `json:"scopes,omitempty"`
	ResourceScopes []string `json:"resource_scopes,omitempty"`
}
