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

