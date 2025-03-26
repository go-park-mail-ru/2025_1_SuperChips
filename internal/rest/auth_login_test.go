package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mock_user "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/user"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
	"go.uber.org/mock/gomock"
)

func TestLoginHandler(t *testing.T) {
	base := tu.Host + "/login"

	type TestCase struct {
		title  string
		method string
		url    string
		body   string

		email    string
		password string
		userId   uint64

		expectLoginUser bool
		returnLoginUser error

		expectGetUserId bool
		returnGetUserId uint64

		expStatus   int
		expResponse string
	}
	tests := []TestCase{
		{
			title:  "Позитивный сценарий",
			method: http.MethodPost,
			url:    base,
			body: tu.Marshal(domain.LoginData{
				Password: "qwerty123",
				Email:    "AlexKvas@mail.ru",
			}),

			email:    "AlexKvas@mail.ru",
			password: "qwerty123",
			userId:   42,

			expectLoginUser: true,
			returnLoginUser: nil,

			expectGetUserId: true,
			returnGetUserId: 42,

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "OK",
			}),
		},
		{
			title:  "Некорректный благополучный сценарий: некорректный запрос (GET вместо POST) -> отрабатывает позитивный сценарий",
			method: http.MethodGet,
			url:    base,
			body: tu.Marshal(domain.LoginData{
				Password: "qwerty123",
				Email:    "AlexKvas@mail.ru",
			}),

			email:    "AlexKvas@mail.ru",
			password: "qwerty123",
			userId:   42,

			expectLoginUser: true,
			returnLoginUser: nil,

			expectGetUserId: true,
			returnGetUserId: 42,

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "OK",
			}),
		},
		{
			title:  "Некорректный сценарий: пустое тело запроса",
			method: http.MethodPost,
			url:    base,
			body:   "",

			email:    "",
			password: "",
			userId:   0,

			expectLoginUser: false,
			returnLoginUser: nil,

			expectGetUserId: false,
			returnGetUserId: 0,

			expStatus: http.StatusBadRequest,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "Bad Request",
			}),
		},
		// TODO: Мистический случай.
		{
			title:  "Некорректный сценарий: email меньше 3 символов",
			method: http.MethodPost,
			url:    base,
			body: tu.Marshal(domain.LoginData{
				Password: "qwerty123",
				Email:    "em",
			}),

			email:    "em",
			password: "qwerty123",
			userId:   42,

			expectLoginUser: false, // Мистика здесь: EXPECT не выполняется, однако метод благополучно вызывается. True поставить нельзя - сломается.
			returnLoginUser: nil,

			expectGetUserId: false,
			returnGetUserId: 0,

			expStatus: http.StatusBadRequest,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "validation failed",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := tu.TestConfig

			mockUserRepo := mock_user.NewMockUserRepository(ctrl)
			mockUserService := user.NewUserService(mockUserRepo)

			if tt.expectLoginUser {
				mockUserRepo.EXPECT().
					LoginUser(tt.email, tt.password).
					Return(tt.returnLoginUser)
			}
			if tt.expectGetUserId {
				mockUserRepo.EXPECT().
					GetUserId(tt.email).
					Return(tt.returnGetUserId)
			}

			app := rest.AuthHandler{
				Config:      cfg,
				UserService: *mockUserService,
				JWTManager:  *auth.NewJWTManager(cfg),
			}

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			app.LoginHandler(rr, req)

			if rr.Code != tt.expStatus {
				tu.PrintDifference(t, "StatusCode", rr.Code, tt.expStatus)
			}

			gotResponse := tu.GetBodyJson(rr)
			if gotResponse != tt.expResponse {
				tu.PrintDifference(t, "Response", gotResponse, tt.expResponse)
			}
		})
	}
}
