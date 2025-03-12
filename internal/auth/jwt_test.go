package auth_test

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
)

func TestCreateJWT(t *testing.T) {
	config := configs.Config{
		JWTSecret:      []byte("secret"),
		ExpirationTime: time.Minute * 10,
	}

	tests := []struct {
		userID uint64
		email  string
		err    error
	}{
		{userID: 1, email: "test@example.com", err: nil},                 // Валидный случай.
		{userID: 0, email: "test@example.com", err: auth.ErrInvalidUser}, // Некорректный userID.
	}

	for i, c := range tests {
		t.Run(c.email, func(t *testing.T) {
			token, err := auth.CreateJWT(config, c.userID, c.email)

			if err != c.err {
				printDifference(t, i, "Error", err, c.err)
			}

			if err == nil && token == "" {
				printDifference(t, i, "Token", token, "non-empty token")
			}
		})
	}
}

func TestParseJWTToken(t *testing.T) {
	config := configs.Config{
		JWTSecret:      []byte("secret"),
		ExpirationTime: time.Minute * 10,
	}

	token, _ := auth.CreateJWT(config, 1, "test@example.com")

	tests := []struct {
		token string
		err   error
	}{
		{token: token, err: nil},                             // Валидный токен.
		{token: "invalidToken", err: auth.ErrorExpiredToken}, // Некорректный токен.
	}

	for i, c := range tests {
		t.Run(c.token, func(t *testing.T) {
			claims, err := auth.ParseJWTToken(c.token, config)

			if err != c.err {
				printDifference(t, i, "Error", err, c.err)
			}

			if c.err != nil {
				if err == nil {
					printDifference(t, i, "Error", err, c.err)
				}
			} else {
				if err != nil {
					printDifference(t, i, "Error", err, nil)
				} else if claims == nil {
					printDifference(t, i, "Claims", claims, "non-nil claims")
				} else {
					if claims.Email != "test@example.com" {
						printDifference(t, i, "Email", claims.Email, "test@example.com")
					}
					if claims.UserID != 1 {
						printDifference(t, i, "UserID", claims.UserID, 1)
					}
				}
			}
		})
	}
}

// Тест на токен с истёкшим сроком.
func TestParseExpiredJWTToken(t *testing.T) {
	config := configs.Config{
		JWTSecret:      []byte("secret"),
		ExpirationTime: time.Millisecond * 1,
	}

	token, _ := auth.CreateJWT(config, 1, "test@example.com")

	time.Sleep(time.Millisecond * 2)

	_, err := auth.ParseJWTToken(token, config)
	if err != auth.ErrorExpiredToken {
		printDifference(t, 0, "Error", err, auth.ErrorExpiredToken)
	}
}

func printDifference(t *testing.T, num int, name string, got any, exp any) {
	t.Errorf("[%d] wrong %v", num, name)
	t.Errorf("--> got     : %+v", got)
	t.Errorf("--> expected: %+v", exp)
}
