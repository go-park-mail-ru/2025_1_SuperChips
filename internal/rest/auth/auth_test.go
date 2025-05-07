package rest

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/stretchr/testify/require"
)

func TestCreateJWT(t *testing.T) {
	validConfig := configs.Config{
		JWTSecret:       []byte("valid-secret-32-chars-long-123456"),
		ExpirationTime: time.Hour,
	}

	tests := []struct {
		name      string
		cfg       configs.Config
		userID    int
		email     string
		username  string
		wantError error
	}{
		{
			name:      "Valid user",
			cfg:       validConfig,
			userID:    1,
			email:     "valid@example.com",
			username:  "cooluser",
			wantError: nil,
		},
		{
			name:      "Invalid user ID 0",
			cfg:       validConfig,
			userID:    0,
			email:     "invalid@example.com",
			username:  "cooluser",
			wantError: ErrInvalidUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mngr := NewJWTManager(tt.cfg)
			token, err := mngr.CreateJWT(tt.email, tt.username, tt.userID)

			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				require.Empty(t, token)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, token)
				
				parsed, err := mngr.ParseJWTToken(token)
				require.NoError(t, err)
				require.Equal(t, tt.userID, parsed.UserID)
				require.Equal(t, tt.email, parsed.Email)
			}
		})
	}
}

func TestParseJWTToken(t *testing.T) {
	validConfig := configs.Config{
		JWTSecret:       []byte("valid-secret-32-chars-long-123456"),
		ExpirationTime: time.Hour,
	}

	expiredConfig := configs.Config{
		JWTSecret:       validConfig.JWTSecret,
		ExpirationTime: -time.Hour,
	}

	invalidSecretConfig := configs.Config{
		JWTSecret:       []byte("different-secret-32-chars-long-456"),
		ExpirationTime: time.Hour,
	}

	validMngr := NewJWTManager(validConfig)
	validToken, _ := validMngr.CreateJWT("valid@example.com", "cooluser", 1)

	expiredMngr := NewJWTManager(expiredConfig)
	expiredToken, _ := expiredMngr.CreateJWT("expired@example.com", "cooluser", 2)

	invalidMngr := NewJWTManager(invalidSecretConfig)
	invalidToken, _ := invalidMngr.CreateJWT("invalid@example.com", "cooluser", 3)

	tests := []struct {
		name        string
		tokenString string
		mngr        *JWTManager
		wantError   error
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			mngr:        validMngr,
			wantError:   nil,
		},
		{
			name:        "Expired token",
			tokenString: expiredToken,
			mngr:        validMngr,
			wantError:   ErrorExpiredToken,
		},
		{
			name:        "Invalid signature",
			tokenString: invalidToken,
			mngr:        validMngr,
			wantError:   ErrorExpiredToken,
		},
		{
			name:        "Malformed token",
			tokenString: "invalid.token.string",
			mngr:        validMngr,
			wantError:   ErrorExpiredToken,
		},
		{
			name:        "Empty token",
			tokenString: "",
			mngr:        validMngr,
			wantError:   ErrorExpiredToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tt.mngr.ParseJWTToken(tt.tokenString)
			
			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				require.Nil(t, claims)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				require.Equal(t, 1, claims.UserID)
				require.Equal(t, "valid@example.com", claims.Email)
			}
		})
	}
}