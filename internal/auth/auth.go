package auth

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
)

func SetCookieJWT(w http.ResponseWriter, config configs.Config, email string, userID uint64) error {
    tokenString, err := CreateJWT(config, userID, email)
    if err != nil {
        return err
    }

    http.SetCookie(w, &http.Cookie{
        Name:     AuthToken,
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   config.CookieSecure,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Now().Add(config.ExpirationTime),
    })

    return nil
}

