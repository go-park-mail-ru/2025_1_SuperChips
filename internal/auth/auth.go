package auth

import (
	"net/http"
	"time"
)

func SetCookieJWT(w http.ResponseWriter, email string) error {
    expirationTime := time.Minute * 15
    tokenString, err := CreateJWT(email, expirationTime)
    if err != nil {
        return err
    }

    http.SetCookie(w, &http.Cookie{
        Name:     AuthToken,
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Now().Add(expirationTime),
    })

    return nil
}