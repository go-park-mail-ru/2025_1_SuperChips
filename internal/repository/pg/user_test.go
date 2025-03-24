package repository_test

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
//     pg "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
// )

// func TestUserRepository_GetUserPublicInfo(t *testing.T) {
//     userMock, err := mock.NewUserMock()
//     if err != nil {
//         t.Fatalf("failed to create mock: %v", err)
//     }
//     defer userMock.Close()

//     expectedEmail := "test@mail.ru"
//     expectedUser := domain.PublicUser{
//         Username: "test",
//         Email:    expectedEmail,
//         Birthday: time.Date(2023, 3, 3, 0, 0, 0, 0, time.UTC),
//     }
//     userMock.GetUserPublicInfo(expectedEmail)

//     repo, err := pg.NewPGUserStorage(userMock.DB)

//     got, err := repo.GetUserPublicInfo(context.Background(), expectedEmail)
//     if err != nil {
//         t.Fatalf("unexpected error: %v", err)
//     }

//     if got.ID != expectedUser.ID {
//         t.Errorf("want ID: %v, got: %v", expectedUser.ID, got.ID)
//     }
//     if got.Email != expectedUser.Email {
//         t.Errorf("want Email: %v, got: %v", expectedUser.Email, got.Email)
//     }
//     if got.Username != expectedUser.Username {
//         t.Errorf("want Username: %v, got: %v", expectedUser.Username, got.Username)
//     }

//     if err := userMock.mock.ExpectationsWereMet(); err != nil {
//         t.Errorf("unmet expectations: %v", err)
//     }
// }