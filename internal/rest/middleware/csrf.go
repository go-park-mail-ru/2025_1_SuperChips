package middleware

import (
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/csrf"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
)

func CSRFMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(csrf.CSRFToken)
			if err != nil {
				rest.HttpErrorToJson(w, "csrf token missing", http.StatusForbidden)
				return
			}

			token := cookie.Value
			requestToken := r.Header.Get("X-CSRF-TOKEN")
			if token != requestToken {
				rest.HttpErrorToJson(w, "csrf token mismatch", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

