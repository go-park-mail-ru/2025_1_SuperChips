package domain_test

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
)

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		wantErr bool
	}{
		{
			name: "Сценарий: корректный",
			user: domain.User{
				Email:    "test@example.com",
				Username: "username",
				Password: "securepassword123",
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "Сценарий: некорректная почта: слишком длинная.",
			user: domain.User{
				Email:    "lalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalala@b.c",
				Username: "username",
				Password: "securepassword123",
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректная почта: некорректный формат.",
			user: domain.User{
				Email:    "invalid-email",
				Username: "username",
				Password: "securepassword123",
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректная имя пользователя: слишком короткое.",
			user: domain.User{
				Email:    "test@example.com",
				Username: "a",
				Password: "securepassword123",
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректный пароль: отсутствует пароль.",
			user: domain.User{
				Email:    "test@example.com",
				Username: "username",
				Password: "",
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректный пароль: слишком длинный.",
			user: domain.User{
				Email:    "test@example.com",
				Username: "username",
				Password: string(make([]byte, 97)),
				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректная дата рождения: дата из будущего.",
			user: domain.User{
				Email:    "test@example.com",
				Username: "username",
				Password: "securepassword123",
				Birthday: time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "Сценарий: некорректная дата рождения: слишком старая дата.",
			user: domain.User{
				Email:    "test@example.com",
				Username: "username",
				Password: "securepassword123",
				Birthday: time.Now().Add(-200 * 365 * 24 * time.Hour), // More than 150 years old
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.ValidateUser()
			if (err != nil) != tt.wantErr {
				tu.PrintDifference(t, "ValidateUser", err, tt.wantErr)
			}
		})
	}
}
