package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	mock_pin "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/pin"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pin"
	tu "github.com/go-park-mail-ru/2025_1_SuperChips/test_utils"
	"go.uber.org/mock/gomock"
)

func TestPinsHandler_FeedHandler(t *testing.T) {
	base := tu.Host + "/feed"

	type TestCase struct {
		title  string
		method string
		url    string
		query  string
		body   string
		page   int

		pageSize       int
		expectMockCall bool
		mockRepoReturn []domain.PinData

		expStatus   int
		expResponse string
	}
	tests := []TestCase{
		{
			title:  "Позитивный сценарий",
			method: http.MethodGet,
			url:    base,
			query:  "page=2",
			body:   "",
			page:   2,

			pageSize:       10,
			expectMockCall: true,
			mockRepoReturn: []domain.PinData{{Header: "1"}, {Header: "2"}},

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Data: []domain.PinData{{Header: "1"}, {Header: "2"}},
			}),
		},
		{
			title:  "Некорректный благополучный сценарий: количество страниц меньше 1 -> возвращается 1-ая страница",
			method: http.MethodGet,
			url:    base,
			query:  "page=0",
			body:   "",
			page:   1,

			pageSize:       10,
			expectMockCall: true,
			mockRepoReturn: []domain.PinData{{Header: "1"}, {Header: "2"}},

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Data: []domain.PinData{{Header: "1"}, {Header: "2"}},
			}),
		},
		{
			title:  "Некорректный благополучный сценарий: количество страниц не удалость распарсить -> возвращается 1-ая страница",
			method: http.MethodGet,
			url:    base,
			query:  "page=qwerty",
			body:   "",
			page:   1,

			pageSize:       10,
			expectMockCall: true,
			mockRepoReturn: []domain.PinData{{Header: "1"}, {Header: "2"}},

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Data: []domain.PinData{{Header: "1"}, {Header: "2"}},
			}),
		},
		{
			title:  "Некорректный благополучный сценарий: количество страниц не задано -> возвращается 1-ая страница",
			method: http.MethodGet,
			url:    base,
			query:  "",
			body:   "",
			page:   1,

			pageSize:       10,
			expectMockCall: true,
			mockRepoReturn: []domain.PinData{{Header: "1"}, {Header: "2"}},

			expStatus: http.StatusOK,
			expResponse: tu.Marshal(rest.ServerResponse{
				Data: []domain.PinData{{Header: "1"}, {Header: "2"}},
			}),
		},
		{
			title:  "Негативный сценарий: на странице с данным номером нет пинов",
			method: http.MethodGet,
			url:    base,
			query:  "page=999",
			body:   "",
			page:   999,

			pageSize:       10,
			expectMockCall: true,
			mockRepoReturn: []domain.PinData{},

			expStatus: http.StatusNotFound,
			expResponse: tu.Marshal(rest.ServerResponse{
				Description: "Not Found",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := tu.TestConfig
			cfg.PageSize = tt.pageSize

			mockPinRepo := mock_pin.NewMockPinRepository(ctrl)
			mockPinService := pin.NewPinService(mockPinRepo)

			if tt.expectMockCall {
				mockPinRepo.EXPECT().
					GetPins(tt.page, tt.pageSize).
					Return(tt.mockRepoReturn)
			}

			app := rest.PinsHandler{
				Config:     cfg,
				PinService: mockPinService,
			}

			req := httptest.NewRequest(tt.method, tt.url+"?"+tt.query, strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			app.FeedHandler(rr, req)

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
