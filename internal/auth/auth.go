package auth

import (
	"net/http"
	"time"
)

func SetAuthCookie(w http.ResponseWriter, tokenString string, expirationTime time.Duration) {
    http.SetCookie(w, &http.Cookie{
        Name:     AuthToken,
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Now().Add(expirationTime),
    })
}

func CookieAddJWT(w http.ResponseWriter, email string) error {
    tokenString, err := CreateJWT(email, 15*time.Minute)
    if err != nil {
        return err
    }

    SetAuthCookie(w, tokenString, 15*time.Minute)

    return nil
}