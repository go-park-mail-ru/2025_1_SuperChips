package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/like/service"
)

func TestLikeFlow_Success(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockLikeService := mocks.NewMockLikeService(ctrl)
    pinID := 456
    userID := 123
    expectedAction := "liked"
    expectedAuthorUsername := "testuser"
    mockLikeService.EXPECT().
        LikeFlow(gomock.Any(), pinID, userID).
        Return(expectedAction, expectedAuthorUsername, nil)

    handler := &LikeHandler{
        NotificationChan: make(chan<- domain.WebMessage, 5),
        LikeService:    mockLikeService,
        ContextTimeout: 5 * time.Second,
    }

    payload := map[string]int{"pin_id": pinID}
    payloadBytes, err := json.Marshal(payload)
    assert.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/like", bytes.NewReader(payloadBytes))
    req.Header.Set("Content-Type", "application/json")

    claims := &auth.Claims{UserID: userID}
    ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
    req = req.WithContext(ctx)

    rr := httptest.NewRecorder()

    handler.LikeFlow(rr, req)

    assert.Equal(t, http.StatusOK, rr.Code)

    var response ServerResponse
    err = json.Unmarshal(rr.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, "OK", response.Description)

    dataBytes, err := json.Marshal(response.Data)
    assert.NoError(t, err)
    var data map[string]interface{}
    err = json.Unmarshal(dataBytes, &data)
    assert.NoError(t, err)
    assert.Equal(t, expectedAction, data["action"])
}

func TestLikeFlow_ServiceError(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockLikeService := mocks.NewMockLikeService(ctrl)
    pinID := 456
    userID := 123
    serviceErr := errors.New("like service failure")
    mockLikeService.EXPECT().
        LikeFlow(gomock.Any(), pinID, userID).
        Return("", "", serviceErr)

    handler := &LikeHandler{
        NotificationChan: make(chan<- domain.WebMessage, 5),
        LikeService:    mockLikeService,
        ContextTimeout: 5 * time.Second,
    }

    payload := map[string]int{"pin_id": pinID}
    payloadBytes, err := json.Marshal(payload)
    assert.NoError(t, err)

    req := httptest.NewRequest(http.MethodPost, "/api/v1/like", bytes.NewReader(payloadBytes))
    req.Header.Set("Content-Type", "application/json")
    claims := &auth.Claims{UserID: userID}
    ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
    req = req.WithContext(ctx)

    rr := httptest.NewRecorder()
    handler.LikeFlow(rr, req)

    assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestLikeFlow_DecodeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLikeService := mocks.NewMockLikeService(ctrl)

	handler := &LikeHandler{
		NotificationChan: make(chan<- domain.WebMessage, 5),
		LikeService:    mockLikeService,
		ContextTimeout: 5 * time.Second,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/like", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.Claims{UserID: 123}
	ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, claims)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.LikeFlow(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
