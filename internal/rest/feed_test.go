package rest_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	mock_pin "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/feed/grpc"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/feed"
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
		mockRepoError  error

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
			mockRepoError:  nil,

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
			mockRepoError:  nil,

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
			mockRepoError:  nil,

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
			mockRepoError:  nil,

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
			mockRepoError:  nil,

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

            // Create a mock FeedClient
            mockGrpcClient := mock_pin.NewMockFeedClient(ctrl)

            // Set up expectations for GetPins if expectMockCall is true
            if tt.expectMockCall {
                mockGrpcClient.EXPECT().
                    GetPins(gomock.Any(), &gen.GetPinsRequest{
                        Page:     int64(tt.page),
                        PageSize: int64(tt.pageSize),
                    }).
                    Return(&gen.GetPinsResponse{
                        Pins: PinsToGrpc(tt.mockRepoReturn),
                    }, tt.mockRepoError)
            }

            // Initialize the PinsHandler with the mock client
            app := rest.PinsHandler{
                Config:           cfg,
                FeedClient:       mockGrpcClient,
                ContextExpiration: time.Hour,
            }

            // Create the HTTP request
            req := httptest.NewRequest(tt.method, tt.url+"?"+tt.query, strings.NewReader(tt.body))
            rr := httptest.NewRecorder()

            // Call the handler
            app.FeedHandler(rr, req)

            // Assert the response status code
            if rr.Code != tt.expStatus {
                tu.PrintDifference(t, "StatusCode", rr.Code, tt.expStatus)
            }

            // Assert the response body
            gotResponse := tu.GetBodyJson(rr)
            if gotResponse != tt.expResponse {
                tu.PrintDifference(t, "Response", gotResponse, tt.expResponse)
            }
        })
    }
}

func PinsToGrpc(pins []domain.PinData) []*gen.Pin {
    var grpcPins []*gen.Pin
    for _, pin := range pins {
        grpcPins = append(grpcPins, &gen.Pin{
            FlowId:         pin.FlowID,
            Header:         pin.Header,
            AuthorId:       pin.AuthorID,
            AuthorUsername: pin.AuthorUsername,
            Description:    pin.Description,
            MediaUrl:       pin.MediaURL,
            IsPrivate:      pin.IsPrivate,
            CreatedAt:      pin.CreatedAt,
            UpdatedAt:      pin.UpdatedAt,
            IsLiked:        pin.IsLiked,
            LikeCount:      int64(pin.LikeCount),
            Width:          int64(pin.Width),
            Height:         int64(pin.Height),
        })
    }
    return grpcPins
}
