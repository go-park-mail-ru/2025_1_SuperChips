package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

// Миделваре для аутентификации.
// Предусмотрен аргумент block для прерывания(true)/продолжения(false) обработки запроса, если в Cookie отсутствуют данные для авторизации.
func AuthMiddleware(jwtManager *auth.JWTManager, block bool) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.AuthToken)
			if err != nil {
				if block {
					rest.HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			token := cookie.Value
			claims, err := jwtManager.ParseJWTToken(token)
			if err != nil {
				if block {
					rest.HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

func AuthSoftMiddleware(jwtManager *auth.JWTManager) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie(auth.AuthToken)
            if err != nil {
                next.ServeHTTP(w, r)
                return
            }

            token := cookie.Value
            claims, err := jwtManager.ParseJWTToken(token)
            if err != nil {
                return
            }   

            ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        }
    }
}
