package entity

import (

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager interface {
	NewJWTManager(cfg configs.Config)
	CreateJWT(email string, userID int) (string, error)
	ParseJWTToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID int
	Email  string
	jwt.RegisteredClaims
}

