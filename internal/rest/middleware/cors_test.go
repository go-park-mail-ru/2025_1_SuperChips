package middleware

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-park-mail-ru/2025_1_SuperChips/configs"
    "github.com/stretchr/testify/assert"
)

func TestCorsMiddleware(t *testing.T) {
    type TestCase struct {
        name             string
        config           configs.Config
        allowedMethods   []string
        requestMethod    string
        requestOrigin    string
        requestHeaders   map[string]string
        expectedStatus   int
        expectedHeaders  map[string]string
        expectNextCalled bool
    }

    tests := []TestCase{
        {
            name: "Valid origin in production",
            config: configs.Config{
                Environment:  "prod",
                AllowedOrigins: []string{"http://example.com"},
            },
            allowedMethods: []string{http.MethodGet},
            requestMethod:  http.MethodGet,
            requestOrigin:  "http://example.com",
            expectedStatus: http.StatusOK,
            expectedHeaders: map[string]string{
                "Access-Control-Allow-Origin":      "http://example.com",
                "Access-Control-Allow-Methods":     "GET",
                "Access-Control-Allow-Headers":     "Content-Type, Authorization, X-CSRF-Token, X-Forwarded-Host",
                "Access-Control-Allow-Credentials": "true",
            },
            expectNextCalled: true,
        },
        {
            name: "Invalid origin in production",
            config: configs.Config{
                Environment:  "prod",
                AllowedOrigins: []string{"http://example.com"},
            },
            allowedMethods: []string{http.MethodGet},
            requestMethod:  http.MethodGet,
            requestOrigin:  "http://invalid.com",
            expectedStatus: http.StatusForbidden,
            expectNextCalled: false,
        },
        {
            name: "X-Forwarded-Host allowed in production",
            config: configs.Config{
                Environment:  "prod",
                AllowedOrigins: []string{"example.com"},
            },
            allowedMethods: []string{http.MethodPost},
            requestMethod:  http.MethodPost,
            requestHeaders: map[string]string{
                "X-Forwarded-Host": "example.com",
            },
            expectedStatus: http.StatusOK,
            expectedHeaders: map[string]string{
                "Access-Control-Allow-Origin":  "example.com",
            },
            expectNextCalled: true,
        },
        {
            name: "OPTIONS request",
            config: configs.Config{
                Environment:  "prod",
                AllowedOrigins: []string{"*"},
            },
            allowedMethods: []string{http.MethodGet, http.MethodOptions},
            requestMethod:  http.MethodOptions,
            requestOrigin:  "http://any.com",
            expectedStatus: http.StatusOK,
            expectedHeaders: map[string]string{
                "Access-Control-Allow-Origin":  "*",
                "Access-Control-Allow-Methods": "GET, OPTIONS",
            },
            expectNextCalled: false,
        },
        {
            name: "Disallowed method",
            config: configs.Config{
                Environment:  "prod",
                AllowedOrigins: []string{"*"},
            },
            allowedMethods: []string{http.MethodGet},
            requestMethod:  http.MethodPost,
            expectedStatus: http.StatusMethodNotAllowed,
            expectNextCalled: false,
        },
        {
            name: "Non-production environment",
            config: configs.Config{
                Environment:  "dev",
            },
            allowedMethods: []string{http.MethodGet},
            requestMethod:  http.MethodGet,
            expectedHeaders: map[string]string{
                "Access-Control-Allow-Origin": "*",
            },
            expectedStatus:   http.StatusOK,
            expectNextCalled: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var nextCalled bool
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                nextCalled = true
            })

            middleware := CorsMiddleware(tt.config, tt.allowedMethods)(handler)

            req := httptest.NewRequest(tt.requestMethod, "/", nil)
            for k, v := range tt.requestHeaders {
                req.Header.Set(k, v)
            }
            req.Header.Set("Origin", tt.requestOrigin)

            rr := httptest.NewRecorder()

            middleware.ServeHTTP(rr, req)

            assert.Equal(t, tt.expectedStatus, rr.Code)
            for k, v := range tt.expectedHeaders {
                assert.Equal(t, v, rr.Header().Get(k))
            }
            assert.Equal(t, tt.expectNextCalled, nextCalled)
        })
    }
}

