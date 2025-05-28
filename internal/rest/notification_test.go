package rest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) GetNotifications(ctx context.Context, userID uint) ([]domain.Notification, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func TestNotificationHandler_GetNotifications(t *testing.T) {
	now := time.Now()
	testNotification := domain.Notification{
		ID:                   1,
		Type:                 "friend_request",
		CreatedAt:            now,
		SenderUsername:       "sender_user",
		SenderAvatar:         "avatar.jpg",
		SenderExternalAvatar: true,
		ReceiverUsername:     "receiver_user",
		IsRead:               false,
		AdditionalData:       map[string]interface{}{"request_id": 123},
	}

	tests := []struct {
		name               string
		userID             int
		mockNotifications  []domain.Notification
		mockError          error
		expectedStatus     int
		expectedResponse   string
	}{
		{
			name:               "Success with notifications",
			userID:             1,
			mockNotifications:  []domain.Notification{testNotification},
			expectedStatus:     http.StatusOK,
			expectedResponse:   mustMarshal(t, testNotification),
		},
		{
			name:               "Empty notifications",
			userID:             2,
			mockNotifications:  []domain.Notification{},
			expectedStatus:     http.StatusOK,
			expectedResponse:   `{"description":"OK","data":[]}`,
		},
		{
			name:               "Not Found",
			userID:             3,
			mockError:          domain.ErrNotFound,
			expectedStatus:     http.StatusNotFound,
			expectedResponse:   "not found\n",
		},
		{
			name:               "Internal Server Error",
			userID:             4,
			mockError:          errors.New("database error"),
			expectedStatus:     http.StatusInternalServerError,
			expectedResponse:   "database error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockNotificationService)
			handler := NotificationHandler{
				NotificationService: mockService,
				ContextExpiration:   time.Second,
			}

			mockService.On("GetNotifications", mock.Anything, uint(tt.userID)).
				Return(tt.mockNotifications, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
			req = req.WithContext(context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: tt.userID}))

			w := httptest.NewRecorder()

			handler.GetNotifications(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			if tt.expectedStatus == http.StatusOK {
				var response ServerResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				if len(tt.mockNotifications) > 0 {
					notifications := response.Data.([]interface{})
					notification := notifications[0].(map[string]interface{})
					
					assert.Equal(t, float64(testNotification.ID), notification["id"].(float64))
					assert.Equal(t, testNotification.Type, notification["type"].(string))
					assert.Equal(t, testNotification.SenderUsername, notification["sender"].(string))
					assert.Equal(t, testNotification.SenderAvatar, notification["sender_avatar"].(string))
					assert.Equal(t, testNotification.ReceiverUsername, notification["receiver"].(string))
					assert.Equal(t, testNotification.IsRead, notification["is_read"].(bool))
					
					// Verify additional data
					additionalData := notification["additional_data"].(map[string]interface{})
					assert.Equal(t, float64(123), additionalData["request_id"].(float64))
				}
			} else {
				assert.Equal(t, tt.expectedResponse, w.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}

func mustMarshal(t *testing.T, n domain.Notification) string {
	resp := ServerResponse{
		Description: "OK",
		Data:        []domain.Notification{n},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal notification: %v", err)
	}
	return string(data)
}

func TestHandleNotificationError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Not Found",
			err:            domain.ErrNotFound,
			expectedStatus: http.StatusNotFound,
			expectedBody:   "not found\n",
		},
		{
			name:           "Internal Server Error",
			err:            errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			handleNotificationError(w, tt.err)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Equal(t, tt.expectedBody, w.Body.String())
		})
	}
}
