package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-park-mail-ru/2025_1_SuperChips/internal/csrf"
    "github.com/stretchr/testify/assert"
)

func TestCSRFMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        cookieValue    string
        headerValue    string
        expectedStatus int
        expectedBody   string
    }{
        {
            name:           "Valid CSRF token",
            cookieValue:    "valid_token",
            headerValue:    "valid_token",
            expectedStatus: http.StatusOK,
            expectedBody:   "",
        },
        {
            name:           "Missing CSRF cookie",
            cookieValue:    "",
            headerValue:    "valid_token",
            expectedStatus: http.StatusForbidden,
            expectedBody:   `{"description":"csrf token missing"}`,
        },
        {
            name:           "CSRF token mismatch",
            cookieValue:    "valid_token",
            headerValue:    "invalid_token",
            expectedStatus: http.StatusForbidden,
            expectedBody:   `{"description":"csrf token mismatch"}`,
        },
        {
            name:           "Missing X-CSRF-TOKEN header",
            cookieValue:    "valid_token",
            headerValue:    "",
            expectedStatus: http.StatusForbidden,
            expectedBody:   `{"description":"csrf token mismatch"}`,
        },
        {
            name:           "Empty token values",
            cookieValue:    "",
            headerValue:    "",
            expectedStatus: http.StatusForbidden,
            expectedBody:   `{"description":"csrf token missing"}`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            })

            handler := CSRFMiddleware()(nextHandler)

            req := httptest.NewRequest("POST", "/test", nil)
            
            if tt.cookieValue != "" {
                req.AddCookie(&http.Cookie{
                    Name:  csrf.CSRFToken,
                    Value: tt.cookieValue,
                })
            }

            req.Header.Set("X-CSRF-TOKEN", tt.headerValue)

            rr := httptest.NewRecorder()
            handler.ServeHTTP(rr, req)

            assert.Equal(t, tt.expectedStatus, rr.Code)

            if tt.expectedBody != "" {
                assert.JSONEq(t, tt.expectedBody, rr.Body.String())
            }
        })
    }
}

