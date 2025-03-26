package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
)

func TestHealthCheckHandler(t *testing.T) {
	base := tu.Host + "/health"

	type TestCase struct {
		title       string
		method      string
		url         string
		requestBody string
		response    string
		statusCode  int
	}
	cases := []TestCase{
		{
			title:       "Позитивный сценарий",
			method:      http.MethodGet,
			url:         base,
			requestBody: "",
			response:    `{"description":"server is up"}`,
			statusCode:  http.StatusOK,
		},
		{
			title:       "Некорректный допустимый сценарий: запрос POST",
			method:      http.MethodPost,
			url:         base,
			requestBody: "",
			response:    `{"description":"server is up"}`,
			statusCode:  http.StatusOK,
		},
		{
			title:       "Некорректный допустимый сценарий: непустое тело запроса",
			method:      http.MethodGet,
			url:         base,
			requestBody: `{"trash":"trash"}`,
			response:    `{"description":"server is up"}`,
			statusCode:  http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.title, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.requestBody))
			rr := httptest.NewRecorder()

			rest.HealthCheckHandler(rr, req)

			if rr.Code != tt.statusCode {
				tu.PrintDifference(t, "StatusCode", rr.Code, tt.statusCode)
			}

			gotResponse := tu.GetBodyJson(rr)
			if gotResponse != tt.response {
				tu.PrintDifference(t, "Response", gotResponse, tt.response)
			}
		})
	}
}
