package models

import (
	"github.com/golang-jwt/jwt"
)

// jwtCustomClaims are custom claims extending default ones.
type JwtCustomClaims struct {
	jwt.StandardClaims
}

type GrantedData struct {
	Retrospect IntrospectSuccess `json:"retrospect" xml:"retrospect"`
	Claims     jwt.MapClaims     `json:"claims" xml:"claims"`
}
