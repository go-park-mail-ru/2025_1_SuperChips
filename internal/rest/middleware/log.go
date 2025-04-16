package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func Log() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			reqID := "req-" + strconv.FormatInt(time.Now().UnixNano(), 10)
			logger := slog.With("request_id", reqID)
	
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
	
			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration", duration,
				"remote_addr", r.RemoteAddr,
			)
		}
	}
}

