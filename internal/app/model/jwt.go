package model

import "github.com/golang-jwt/jwt/v5"

type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint `json:"id_user"`
	IsModerator bool `json:"is_admin"`
}
