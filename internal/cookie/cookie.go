package cookie

import (
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
)

func CookieAddJWT(w http.ResponseWriter, email string) error {
    tokenString, err := auth.CreateJWT(email, 15*time.Minute)
    if err != nil {
        return err
    }

    auth.SetAuthCookie(w, tokenString, 15*time.Minute)

    return nil
}

