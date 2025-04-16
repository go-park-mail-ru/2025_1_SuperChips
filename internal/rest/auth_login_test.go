package rest_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mock_user "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/user/service"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
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

		expectLoginUser  bool
		returnLoginError error

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

			expectLoginUser:  true,
			returnLoginError: nil,

			expStatus:   http.StatusOK,
			expResponse: `{"description":"OK","data":{"csrf_token":".*"}}`,
		},
		{
			title:  "Некорректный сценарий: GET вместо POST",
			method: http.MethodGet,
			url:    base,
			body: tu.Marshal(domain.LoginData{
				Password: "qwerty123",
				Email:    "AlexKvas@mail.ru",
			}),
			email:    "AlexKvas@mail.ru",
			password: "qwerty123",
			userId:   42,

			expectLoginUser:  true,
			returnLoginError: nil,

			expStatus:   http.StatusOK,
			expResponse: `{"description":"OK","data":{"csrf_token":".*"}}`,
		},
		{
			title:  "Некорректный сценарий: пустое тело",
			method: http.MethodPost,
			url:    base,
			body:   "",

			expectLoginUser:  false,
			returnLoginError: nil,

			expStatus:   http.StatusBadRequest,
			expResponse: `{"description":"Bad Request"}`,
		},
		{
			title:  "Некорректный сценарий: email < 3 символов",
			method: http.MethodPost,
			url:    base,
			body: tu.Marshal(domain.LoginData{
				Password: "qwerty123",
				Email:    "em",
			}),

			email: "em",
			password: "qwerty123",

			expectLoginUser:  true,
			returnLoginError: nil,

			expStatus:   http.StatusInternalServerError,
			expResponse: `{"description":"Internal Server Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := tu.TestConfig
			mockUserService := mock_user.NewMockUserUsecaseInterface(ctrl)

			if tt.expectLoginUser {
				mockUserService.EXPECT().
					LoginUser(tt.email, tt.password).
					Return(tt.userId, tt.returnLoginError)
			}

			app := rest.AuthHandler{
				Config:      cfg,
				UserService: mockUserService,
				JWTManager:  *auth.NewJWTManager(cfg),
			}

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			app.LoginHandler(rr, req)

			if rr.Code != tt.expStatus {
				tu.PrintDifference(t, "StatusCode", rr.Code, tt.expStatus)
			}

			gotResponse := tu.GetBodyJson(rr)
			matched, err := regexp.MatchString(tt.expResponse, gotResponse)
			if err != nil {
				t.Fatalf("Invalid regex pattern: %v", err)
			}
			if !matched {
				tu.PrintDifference(t, "Response", gotResponse, tt.expResponse)
			}

			if tt.expStatus == http.StatusOK {
				foundCookie := false
				for _, c := range rr.Result().Cookies() {
					if c.Name == "auth_token" {
						foundCookie = true
						break
					}
				}
				if !foundCookie {
					t.Error("Expected auth_token cookie to be set")
				}
			}
		})
	}
}
