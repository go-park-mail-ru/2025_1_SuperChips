package rest

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	mocks "github.com/go-park-mail-ru/2025_1_SuperChips/mocks/search/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSearchPins(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearchService := mocks.NewMockSearchService(ctrl)

	handler := SearchHandler{
		Service:        mockSearchService,
		ContextTimeout: time.Second,
	}

	t.Run("Success", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchPins(gomock.Any(), query, page, pageSize).
			Return([]domain.PinData{
				{Header: "Pin 1", MediaURL: "image1.jpg"},
				{Header: "Pin 2", MediaURL: "image2.jpg"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/pins?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchPins(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"header":"Pin 1"`)
		assert.Contains(t, rr.Body.String(), `"media_url":"image1.jpg"`)
	})

	t.Run("Validation Error - Missing Query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/pins?page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchPins(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "cannot be empty")
	})

	t.Run("Empty Results", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchPins(gomock.Any(), query, page, pageSize).
			Return([]domain.PinData{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/pins?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchPins(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Not Found")
	})

	t.Run("Database Error", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchPins(gomock.Any(), query, page, pageSize).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/pins?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchPins(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})
}

func TestSearchBoards(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearchService := mocks.NewMockSearchService(ctrl)

	handler := SearchHandler{
		Service:        mockSearchService,
		ContextTimeout: time.Second,
	}

	t.Run("Success", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchBoards(gomock.Any(), query, page, pageSize).
			Return([]domain.Board{
				{Name: "Board 1"},
				{Name: "Board 2"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/boards?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchBoards(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"name":"Board 1"`)
	})

	t.Run("Validation Error - Missing Query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/boards?page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchBoards(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "cannot be empty")
	})

	t.Run("Empty Results", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchBoards(gomock.Any(), query, page, pageSize).
			Return([]domain.Board{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/boards?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchBoards(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Not Found")
	})

	t.Run("Database Error", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchBoards(gomock.Any(), query, page, pageSize).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/boards?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchBoards(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})
}

func TestSearchUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSearchService := mocks.NewMockSearchService(ctrl)

	handler := SearchHandler{
		Service:        mockSearchService,
		ContextTimeout: time.Second,
	}

	t.Run("Success", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchUsers(gomock.Any(), query, page, pageSize).
			Return([]domain.PublicUser{
				{Username: "user1", PublicName: "User One"},
				{Username: "user2", PublicName: "User Two"},
			}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchUsers(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"username":"user1"`)
		assert.Contains(t, rr.Body.String(), `"public_name":"User One"`)
	})

	t.Run("Validation Error - Missing Query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchUsers(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "cannot be empty")
	})

	t.Run("Empty Results", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchUsers(gomock.Any(), query, page, pageSize).
			Return([]domain.PublicUser{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchUsers(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Contains(t, rr.Body.String(), "Not Found")
	})

	t.Run("Database Error", func(t *testing.T) {
		query := "kittens"
		page := 1
		pageSize := 10

		mockSearchService.EXPECT().
			SearchUsers(gomock.Any(), query, page, pageSize).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?query=kittens&page=1&size=10", nil)
		rr := httptest.NewRecorder()

		handler.SearchUsers(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Internal Server Error")
	})
}
