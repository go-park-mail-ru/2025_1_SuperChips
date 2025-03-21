package domain_test

// import (
// 	"testing"
// 	"time"

// 	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/map"
// )

// func TestValidateUser(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		user    entity.User
// 		wantErr bool
// 	}{
// 		{
// 			name: "Сценарий: корректный",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "username",
// 				Password: "securepassword123",
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			name: "Сценарий: некорректная почта: слишком длинная.",
// 			user: entity.User{
// 				Email:    "lalalalalalalalalalalalalalalalalalalalalalalalalalalalalalalala@b.c",
// 				Username: "username",
// 				Password: "securepassword123",
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректная почта: некорректный формат.",
// 			user: entity.User{
// 				Email:    "invalid-email",
// 				Username: "username",
// 				Password: "securepassword123",
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректная имя пользователя: слишком короткое.",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "a",
// 				Password: "securepassword123",
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректный пароль: отсутствует пароль.",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "username",
// 				Password: "",
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректный пароль: слишком длинный.",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "username",
// 				Password: string(make([]byte, 97)),
// 				Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректная дата рождения: дата из будущего.",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "username",
// 				Password: "securepassword123",
// 				Birthday: time.Now().Add(1 * time.Hour),
// 			},
// 			wantErr: true,
// 		},
// 		{
// 			name: "Сценарий: некорректная дата рождения: слишком старая дата.",
// 			user: entity.User{
// 				Email:    "test@example.com",
// 				Username: "username",
// 				Password: "securepassword123",
// 				Birthday: time.Now().Add(-200 * 365 * 24 * time.Hour), // More than 150 years old
// 			},
// 			wantErr: true,
// 		},
// 	}

// 	for i, c := range tests {
// 		t.Run(c.name, func(t *testing.T) {
// 			err := c.user.ValidateUser()
// 			if (err != nil) != c.wantErr {
// 				printDifference(t, i, "ValidateUser", err, c.wantErr)
// 			}
// 		})
// 	}
// }

// func TestMapUserStorage_AddUser(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	// Создаём пользователя
// 	user := entity.User{
// 		Username: "testUser",
// 		Email:    "test@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}

// 	err := storage.AddUser(user)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	storedUser, err := storage.GetUserPublicInfo(user.Email)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}
// 	if storedUser.Username != user.Username || storedUser.Email != user.Email {
// 		t.Fatalf("expected user %v, got %v", user, storedUser)
// 	}
// }

// func TestMapUserStorage_AddUser_WithExistingEmail(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	user1 := entity.User{
// 		Username: "user1",
// 		Email:    "user@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err := storage.AddUser(user1)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	user2 := entity.User{
// 		Username: "user2",
// 		Email:    "user@example.com",
// 		Password: "password456",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err = storage.AddUser(user2)
// 	if err == nil {
// 		t.Fatalf("expected error, got none")
// 	}
// }

// func TestMapUserStorage_AddUser_UsernameAlreadyTaken(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	user1 := entity.User{
// 		Username: "existingUser",
// 		Email:    "user1@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err := storage.AddUser(user1)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	user2 := entity.User{
// 		Username: "existingUser",
// 		Email:    "user2@example.com",
// 		Password: "password456",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}

// 	err = storage.AddUser(user2)
// 	if err == nil {
// 		t.Fatalf("expected error, got nil")
// 	}

// 	expectedErrorMessage := "resource conflict: the username is already used"
// 	if err.Error() != expectedErrorMessage {
// 		t.Fatalf("expected error message '%s', got '%s'", expectedErrorMessage, err.Error())
// 	}
// }

// func TestMapUserStorage_LoginUser(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	user := entity.User{
// 		Username: "testUser",
// 		Email:    "test@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err := storage.AddUser(user)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	err = storage.LoginUser(user.Email, user.Password)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	err = storage.LoginUser(user.Email, "wrongPassword")
// 	if err == nil {
// 		t.Fatalf("expected error, got none")
// 	}
// }

// func TestMapUserStorage_LoginUser_UserNotFound(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	err := storage.LoginUser("nonexistent@example.com", "somePassword")

// 	if err == nil {
// 		t.Fatalf("expected error, got nil")
// 	}

// 	expectedErrorMessage := "invalid credentials: invalid credentials"
// 	if err.Error() != expectedErrorMessage {
// 		t.Fatalf("expected error message '%s', got '%s'", expectedErrorMessage, err.Error())
// 	}
// }

// func TestMapUserStorage_GetUserPublicInfo(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	user := entity.User{
// 		Username: "testUser",
// 		Email:    "test@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err := storage.AddUser(user)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	publicInfo, err := storage.GetUserPublicInfo(user.Email)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	if publicInfo.Username != user.Username || publicInfo.Email != user.Email {
// 		t.Fatalf("expected public info %v, got %v", user, publicInfo)
// 	}

// 	_, err = storage.GetUserPublicInfo("nonexistent@example.com")
// 	if err == nil {
// 		t.Fatalf("expected error, got none")
// 	}
// }

// func TestMapUserStorage_GetUserId(t *testing.T) {
// 	storage := repository.MapUserStorage{}
// 	storage.NewStorage()

// 	user := entity.User{
// 		Username: "testUser",
// 		Email:    "test@example.com",
// 		Password: "password123",
// 		Birthday: time.Date(1990, time.May, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	err := storage.AddUser(user)
// 	if err != nil {
// 		t.Fatalf("expected no error, got %v", err)
// 	}

// 	userId := storage.GetUserId(user.Email)
// 	if userId == 0 {
// 		t.Fatalf("expected non-zero user ID, got %v", userId)
// 	}

// 	userId = storage.GetUserId("nonexistent@example.com")
// 	if userId != 0 {
// 		t.Fatalf("expected zero user ID, got %v", userId)
// 	}
// }

// func printDifference(t *testing.T, num int, name string, got any, exp any) {
// 	t.Errorf("[%d] wrong %v", num, name)
// 	t.Errorf("--> got     : %+v", got)
// 	t.Errorf("--> expected: %+v", exp)
// }
