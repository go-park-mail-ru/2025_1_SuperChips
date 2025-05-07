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

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/subscription/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetUserFollowers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubscriptionService := mocks.NewMockSubscriptionService(ctrl)

	handler := SubscriptionHandler{
		ContextExpiration:   time.Second,
		SubscriptionService: mockSubscriptionService,
	}

	t.Run("Success", func(t *testing.T) {
		followers := []domain.PublicUser{
			{Username: "user1", PublicName: "User One"},
			{Username: "user2", PublicName: "User Two"},
		}

		mockSubscriptionService.EXPECT().
			GetUserFollowers(gomock.Any(), 42, 1, 10).
			Return(followers, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/followers?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowers(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"username":"user1"`)
		assert.Contains(t, rr.Body.String(), `"public_name":"User One"`)
	})

	t.Run("Validation Error - Missing Page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/followers?size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowers(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "page is not specified")
	})

	t.Run("Empty Results", func(t *testing.T) {
		mockSubscriptionService.EXPECT().
			GetUserFollowers(gomock.Any(), 42, 1, 10).
			Return([]domain.PublicUser{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/followers?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowers(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "[]")
	})

	t.Run("Database Error", func(t *testing.T) {
		mockSubscriptionService.EXPECT().
			GetUserFollowers(gomock.Any(), 42, 1, 10).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/followers?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowers(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})
}

func TestGetUserFollowing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubscriptionService := mocks.NewMockSubscriptionService(ctrl)

	handler := SubscriptionHandler{
		ContextExpiration:   time.Second,
		SubscriptionService: mockSubscriptionService,
	}

	t.Run("Success", func(t *testing.T) {
		following := []domain.PublicUser{
			{Username: "user1", PublicName: "User One"},
			{Username: "user2", PublicName: "User Two"},
		}

		mockSubscriptionService.EXPECT().
			GetUserFollowing(gomock.Any(), 42, 1, 10).
			Return(following, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/following?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowing(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"username":"user1"`)
		assert.Contains(t, rr.Body.String(), `"public_name":"User One"`)
	})

	t.Run("Validation Error - Missing Page", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/following?size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowing(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "page is not specified")
	})

	t.Run("Empty Results", func(t *testing.T) {
		mockSubscriptionService.EXPECT().
			GetUserFollowing(gomock.Any(), 42, 1, 10).
			Return([]domain.PublicUser{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/following?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowing(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "[]")
	})

	t.Run("Database Error", func(t *testing.T) {
		mockSubscriptionService.EXPECT().
			GetUserFollowing(gomock.Any(), 42, 1, 10).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/profile/following?page=1&size=10", nil)
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.GetUserFollowing(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})
}

func TestCreateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubscriptionService := mocks.NewMockSubscriptionService(ctrl)

	handler := SubscriptionHandler{
		ContextExpiration:   time.Second,
		SubscriptionService: mockSubscriptionService,
	}

	t.Run("Success", func(t *testing.T) {
		subData := SubscriptionData{TargetUsername: "target_user"}

		mockSubscriptionService.EXPECT().
			CreateSubscription(gomock.Any(), "current_user", "target_user", 42).
			Return(nil)

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.CreateSubscription(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("Validation Error - Missing Target User", func(t *testing.T) {
		subData := SubscriptionData{}

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.CreateSubscription(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Bad Request")
	})

	t.Run("Conflict Error", func(t *testing.T) {
		subData := SubscriptionData{TargetUsername: "target_user"}

		mockSubscriptionService.EXPECT().
			CreateSubscription(gomock.Any(), "current_user", "target_user", 42).
			Return(domain.ErrConflict)

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.CreateSubscription(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "Conflict")
	})
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubscriptionService := mocks.NewMockSubscriptionService(ctrl)

	handler := SubscriptionHandler{
		ContextExpiration:   time.Second,
		SubscriptionService: mockSubscriptionService,
	}

	t.Run("Success", func(t *testing.T) {
		subData := SubscriptionData{TargetUsername: "target_user"}

		mockSubscriptionService.EXPECT().
			DeleteSubscription(gomock.Any(), "target_user", 42).
			Return(nil)

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.DeleteSubscription(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Validation Error - Missing Target User", func(t *testing.T) {
		subData := SubscriptionData{}

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.DeleteSubscription(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Bad Request")
	})

	t.Run("Not Found Error", func(t *testing.T) {
		subData := SubscriptionData{TargetUsername: "target_user"}

		mockSubscriptionService.EXPECT().
			DeleteSubscription(gomock.Any(), "target_user", 42).
			Return(domain.ErrNotFound)

		body, _ := json.Marshal(subData)
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/subscription", bytes.NewBuffer(body))
		ctx := context.WithValue(req.Context(), auth.ClaimsContextKey, &auth.Claims{UserID: 42, Username: "current_user"})
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		handler.DeleteSubscription(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Not Found")
	})
}
