package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
    cfg := configs.Config{
        JWTSecret:     []byte("test-secret"),
        ExpirationTime: 1 * time.Hour,
    }
    jwtManager := auth.NewJWTManager(cfg)

    validToken, err := jwtManager.CreateJWT("test@example.com", "hi", 123)
    assert.NoError(t, err)

    expiredToken, err := jwtManager.CreateJWT("test@example.com", "username", 456)
    assert.NoError(t, err)
    token, _ := jwt.ParseWithClaims(expiredToken, &auth.Claims{}, func(t *jwt.Token) (interface{}, error) {
        return cfg.JWTSecret, nil
    })
    claims := token.Claims.(*auth.Claims)
    claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))
    expiredToken, _ = token.SignedString(cfg.JWTSecret)

    tests := []struct {
        name           string
        block          bool
        cookieValue    string
        expectedStatus int
        expectNext     bool
        expectClaims   *auth.Claims
    }{
        {
            name:           "valid token, block=true",
            block:          true,
            cookieValue:    validToken,
            expectedStatus: http.StatusOK,
            expectNext:     true,
            expectClaims: &auth.Claims{
                UserID: 123,
                Email:  "test@example.com",
                RegisteredClaims: jwt.RegisteredClaims{
                    Issuer: "flow",
                },
            },
        },
        {
            name:           "expired token, block=true",
            block:          true,
            cookieValue:    expiredToken,
            expectedStatus: http.StatusUnauthorized,
            expectNext:     false,
        },
        {
            name:           "invalid token, block=true",
            block:          true,
            cookieValue:    "invalid-token",
            expectedStatus: http.StatusUnauthorized,
            expectNext:     false,
        },
        {
            name:           "missing cookie, block=true",
            block:          true,
            expectedStatus: http.StatusUnauthorized,
            expectNext:     false,
        },
        {
            name:           "valid token, block=false",
            block:          false,
            cookieValue:    validToken,
            expectedStatus: http.StatusOK,
            expectNext:     true,
            expectClaims: &auth.Claims{
                UserID: 123,
                Email:  "test@example.com",
                RegisteredClaims: jwt.RegisteredClaims{
                    Issuer: "flow",
                },
            },
        },
        {
            name:           "expired token, block=false",
            block:          false,
            cookieValue:    expiredToken,
            expectedStatus: http.StatusOK,
            expectNext:     true,
            expectClaims:   nil,
        },
        {
            name:           "invalid token, block=false",
            block:          false,
            cookieValue:    "invalid-token",
            expectedStatus: http.StatusOK,
            expectNext:     true,
            expectClaims:   nil,
        },
        {
            name:           "missing cookie, block=false",
            block:          false,
            expectedStatus: http.StatusOK,
            expectNext:     true,
            expectClaims:   nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var (
                nextCalled bool
                ctxClaims  *auth.Claims
            )

            nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                nextCalled = true
                if claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims); ok {
                    ctxClaims = claims
                } else {
                    ctxClaims = nil
                }
                w.WriteHeader(http.StatusOK)
            })

            middleware := AuthMiddleware(jwtManager, tt.block)(nextHandler)

            req := httptest.NewRequest("GET", "/", nil)
            if tt.cookieValue != "" {
                req.AddCookie(&http.Cookie{Name: auth.AuthToken, Value: tt.cookieValue})
            }
            rr := httptest.NewRecorder()

            middleware.ServeHTTP(rr, req)

            assert.Equal(t, tt.expectedStatus, rr.Code)
            assert.Equal(t, tt.expectNext, nextCalled)

            if tt.expectClaims != nil {
                assert.NotNil(t, ctxClaims)
                assert.Equal(t, tt.expectClaims.UserID, ctxClaims.UserID)
                assert.Equal(t, tt.expectClaims.Email, ctxClaims.Email)
                assert.Equal(t, tt.expectClaims.Issuer, ctxClaims.Issuer)
            } else {
                assert.Nil(t, ctxClaims)
            }
        })
    }
}

