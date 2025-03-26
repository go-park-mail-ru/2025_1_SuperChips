package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mock_user "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/user"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
	"go.uber.org/mock/gomock"
)

func TestLogoutHandler(t *testing.T) {
	base := tu.Host + "/logout"

	type TestCase struct {
		title  string
		method string
		url    string
		body   string

		expStatus   int
		expResponse string
	}
	tests := []TestCase{
		{
			title:  "Позитивный сценарий",
			method: http.MethodPost,
			url:    base,
			body:   "",

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "logged out",
			}),
		},
		{
			title:  "Некорректный сценарий: неверный запрос (GET)",
			method: http.MethodGet,
			url:    base,
			body:   "",

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "logged out",
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

			app := rest.AuthHandler{
				Config:      cfg,
				UserService: mockUserService,
				JWTManager:  *auth.NewJWTManager(cfg),
			}

			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			app.LogoutHandler(rr, req)

			if rr.Code != tt.expStatus {
				tu.PrintDifference(t, "StatusCode", rr.Code, tt.expStatus)
			}

			gotResponse := tu.GetBodyJson(rr)
			if gotResponse != tt.expResponse {
				tu.PrintDifference(t, "Response", gotResponse, tt.expResponse)
			}

			for _, c := range rr.Result().Cookies() {
				if c.Name == "auth_token" {
					if !c.Expires.Before(time.Now()) {
						tu.PrintDifference(t, "Cookie expiration", c.Expires, "less then 0")
					}
					break
				}
			}
		})
	}
}
