package rest

import (
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
)

func CorsMiddleware(next http.HandlerFunc, cfg configs.Config, allowedMethods []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if !slices.Contains(allowedMethods, r.Method) {
			rest.HttpErrorToJson(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		allowedOrigins := cfg.AllowedOrigins
		if cfg.Environment == "prod" {
			origin := r.Header.Get("Origin")
			xForwardedHost := r.Header.Get("X-Forwarded-Host")
			if slices.Contains(allowedOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if slices.Contains(allowedOrigins, xForwardedHost) {
				w.Header().Set("Access-Control-Allow-Origin", xForwardedHost)
			} else {
				rest.HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
