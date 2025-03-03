package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID int
	Email  string
	jwt.RegisteredClaims
}

const AuthToken = "auth_token"

var (
    ErrInvalidUser = errors.New("invalid user")
    ErrSigningJWT  = errors.New("failed to sign JWT")
	ErrorCookieCreation = errors.New("error creating a jwt cookie")
	ErrorJWTParse = errors.New("error parsing jwt token")
	ErrorExpiredToken = errors.New("expired token")
)

var Config configs.Config = configs.LoadConfigFromEnv()

func CreateJWT(email string, expirationTime time.Duration) (string, error) {
    userID := user.GetUserId(email)
    if userID == 0 {
        return "", ErrInvalidUser
    }

    expiration := time.Now().Add(expirationTime)
    claims := &Claims{
        UserID: int(userID),
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expiration),
            Issuer:    "flow",
            ID:        uuid.New().String(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(Config.JWTSecret)
    if err != nil {
        return "", ErrSigningJWT
    }

    return tokenString, nil
}

func ParseJWTToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return Config.JWTSecret, nil
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

