package mtbulkrestapi

import (
	"github.com/dgrijalva/jwt-go"
)

// TokenClaims repesents JWT authentication claims.
type TokenClaims struct {
	AllowedHostPatterns []string
	jwt.StandardClaims
}

// Authenticate represents authenticatation rules.
type Authenticate struct {
	Key                 string   `toml:"key" yaml:"key"`
	AllowedHostPatterns []string `toml:"allowed_host_patterns" yaml:"allowed_host_patterns"`
}
