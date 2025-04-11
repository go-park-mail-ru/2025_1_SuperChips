package middleware

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

func AuthMiddleware(jwtManager *auth.JWTManager) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie(auth.AuthToken)
            if err != nil {
                rest.HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
                return
            }

            token := cookie.Value
            claims, err := jwtManager.ParseJWTToken(token)
            if err != nil {
                rest.HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
                rest.HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
                return
            }

            token := cookie.Value
            claims, err := jwtManager.ParseJWTToken(token)
            if err != nil {
                claims = nil
            }

            ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        }
    }
}
