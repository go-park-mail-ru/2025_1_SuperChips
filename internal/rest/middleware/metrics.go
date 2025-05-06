package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/metrics"
)

func MetricsMiddleware(m metrics.MetricsServicer) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ws := NewResponseWriterStatus(w)
			duration := time.Since(time.Now())

			next.ServeHTTP(ws, r)

			normalizedPath := normalizePath(r.URL.Path)

			m.AddDurationToHistogram(r.Method, normalizedPath, duration)
			m.IncreaseHits(r.Method, normalizedPath, strconv.Itoa(ws.statusCode))
			if ws.statusCode/100 != 2 {
				m.IncreaseErr(r.Method, normalizedPath, ws.description)
			}
		}
	}
}

type ResponseWriterStatus struct {
	http.ResponseWriter
	statusCode  int
	description string
}

func NewResponseWriterStatus(w http.ResponseWriter) *ResponseWriterStatus {
	return &ResponseWriterStatus{w, http.StatusOK, "OK"}
}

func (rw *ResponseWriterStatus) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (w *ResponseWriterStatus) Write(data []byte) (int, error) {
	w.description = ""
	var body struct {
		Description string `json:"description"`
	}
	err := json.Unmarshal(data, &body)
	if err == nil {
		w.description = body.Description
	}
	return w.ResponseWriter.Write(data)
}

func normalizePath(path string) string {
	re := regexp.MustCompile(`/\d+`)
	return re.ReplaceAllString(path, "/:id")
}
