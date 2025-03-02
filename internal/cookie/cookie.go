package cookie

import (
	"errors"
	"fmt"
	"net/http"
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

var (
	ErrorCookieCreation = errors.New("error creating a jwt cookie")
	ErrorJWTParse = errors.New("error parsing jwt token")
	ErrorExpiredToken = errors.New("expired token")
)

var Config configs.Config = configs.LoadConfigFromEnv()

func CookieSetJWT(w http.ResponseWriter, email string) error {
	var err error = nil

	expiration := time.Now().Add(time.Minute * 15)
	claims := &Claims{
		UserID: int(user.GetUserId(email)),
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			Issuer: "flow",
			ID: uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(Config.JWTSecret)
	if err != nil {
		return ErrorCookieCreation
	}

	http.SetCookie(w, &http.Cookie{
		Name: "auth_token",
		Value: tokenString,
		Path: "/",
		HttpOnly: true,
		Secure: false,
		SameSite: http.SameSiteStrictMode,
		Expires: expiration,
	})

	return err
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

