package rest

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const AuthToken = "auth_token"

var (
	ErrInvalidUser    = errors.New("invalid user")
	ErrSigningJWT     = errors.New("failed to sign JWT")
	ErrorJWTParse     = errors.New("error parsing jwt token")
	ErrorExpiredToken = errors.New("expired token")
)

type Claims struct {
	UserID   int
	Username string
	Email    string
	jwt.RegisteredClaims
}

type JWTManager struct {
	secret     []byte
	expiration time.Duration
	issuer     string
}

func NewJWTManager(cfg configs.Config) *JWTManager {
	return &JWTManager{
		secret:     cfg.JWTSecret,
		expiration: cfg.ExpirationTime,
		issuer:     "flow",
	}
}

func (mngr *JWTManager) CreateJWT(email, username string, userID int) (string, error) {
	if userID == 0 {
		return "", ErrInvalidUser
	}

	expiration := time.Now().Add(mngr.expiration)
	claims := &Claims{
		UserID: int(userID),
		Email:  email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			Issuer:    mngr.issuer,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(mngr.secret)
	if err != nil {
		return "", ErrSigningJWT
	}

	return tokenString, nil
}

func (mngr *JWTManager) ParseJWTToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mngr.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrorExpiredToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrorJWTParse
	}

	return claims, nil
}
