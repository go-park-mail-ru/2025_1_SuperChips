package middleware

import (
    "bytes"
	"log/slog"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestLogMiddleware(t *testing.T) {
    var buf bytes.Buffer

    jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    logger := slog.New(jsonHandler)
    slog.SetDefault(logger)

    middleware := Log()

    nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    handlerWithMiddleware := middleware(nextHandler)

    req := httptest.NewRequest("GET", "/test-endpoint", nil)
    req.RemoteAddr = "127.0.0.1:12345"

    rr := httptest.NewRecorder()

    handlerWithMiddleware.ServeHTTP(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code, "Handler returned wrong status code")

    var logEntry map[string]interface{}
    err := json.Unmarshal(buf.Bytes(), &logEntry)
    assert.NoError(t, err, "Failed to parse log entry")

    requiredFields := []string{"request_id", "method", "path", "duration", "remote_addr", "time", "level", "msg"}
    for _, field := range requiredFields {
        assert.Contains(t, logEntry, field, "Missing log field: %s", field)
    }

    assert.Equal(t, "GET", logEntry["method"], "Incorrect HTTP method logged")
    assert.Equal(t, "/test-endpoint", logEntry["path"], "Incorrect path logged")
    assert.Equal(t, "127.0.0.1:12345", logEntry["remote_addr"], "Incorrect remote address logged")
    assert.True(t, strings.HasPrefix(logEntry["request_id"].(string), "req-"), "Invalid request ID format")

    duration, ok := logEntry["duration"].(float64)
    assert.True(t, ok && duration > 0, "Duration should be a positive number")

    assert.Equal(t, "HTTP request", logEntry["msg"], "Unexpected log message")
    assert.Equal(t, "INFO", logEntry["level"], "Log level should be INFO")
}

