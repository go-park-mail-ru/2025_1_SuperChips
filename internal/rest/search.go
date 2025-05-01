package rest

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

type SearchService interface {
	SearchPins(ctx context.Context, query string, page, pageSize int) ([]domain.PinData, error)
	SearchUsers(ctx context.Context, query string, page, pageSize int) ([]domain.PublicUser, error)
	SearchBoards(ctx context.Context, query string, page, pageSize int) ([]domain.Board, error) 
}

type SearchHandler struct {
	Service SearchService
	ContextTimeout time.Duration
}

func (s *SearchHandler) SearchPins(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	query := r.URL.Query().Get("query")
	v.Check(query != "", "query", "cannot be empty")

	page := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		HttpErrorToJson(w, "invalid page", http.StatusBadRequest)
		return
	}

	pageSize := r.URL.Query().Get("size")
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		HttpErrorToJson(w, "invalid size", http.StatusBadRequest)
		return
	}

	v.Check(pageInt >= 0, "page", "cannot be less or equal to zero")

	v.Check(pageSizeInt >= 0 && pageSizeInt <= 30, "page size", "cannot be less than 1 or more than 30")

	if !v.Valid() {
		handleValidatorError(w, v.Errors, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	pins, err := s.Service.SearchPins(ctx, query, pageInt, pageSizeInt)
	if err != nil {
		handleSearchError(w, err)
		return
	}

	if len(pins) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: pins,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (s *SearchHandler) SearchBoards(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	query := r.URL.Query().Get("query")
	v.Check(query != "", "query", "cannot be empty")

	page := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		HttpErrorToJson(w, "invalid page", http.StatusBadRequest)
		return
	}

	pageSize := r.URL.Query().Get("size")
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		HttpErrorToJson(w, "invalid size", http.StatusBadRequest)
		return
	}

	v.Check(pageInt >= 0, "page", "cannot be less or equal to zero")
	v.Check(pageSizeInt >= 0 && pageSizeInt <= 30, "page size", "cannot be less than 1 or more than 30")

	if !v.Valid() {
		handleValidatorError(w, v.Errors, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	boards, err := s.Service.SearchBoards(ctx, query, pageInt, pageSizeInt)
	if err != nil {
		handleSearchError(w, err)
		return
	}

	if len(boards) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: boards,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (s *SearchHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	v := validator.New()

	query := r.URL.Query().Get("query")
	v.Check(query != "", "query", "cannot be empty")

	page := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		HttpErrorToJson(w, "invalid page", http.StatusBadRequest)
		return
	}

	pageSize := r.URL.Query().Get("size")
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		HttpErrorToJson(w, "invalid size", http.StatusBadRequest)
		return
	}

	v.Check(pageInt >= 0, "page", "cannot be less or equal to zero")
	v.Check(pageSizeInt >= 0 && pageSizeInt <= 30, "page size", "cannot be less than 1 or more than 30")

	if !v.Valid() {
		handleValidatorError(w, v.Errors, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.ContextTimeout)
	defer cancel()

	users, err := s.Service.SearchUsers(ctx, query, pageInt, pageSizeInt)
	if err != nil {
		handleSearchError(w, err)
		return
	}

	if len(users) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: users,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleValidatorError(w http.ResponseWriter, errors map[string]string, description string, statusCode int) {
	resp := ServerResponse{
		Description: description,
		Data: errors,
	}

	ServerGenerateJSONResponse(w, resp, statusCode)
}

func handleSearchError(w http.ResponseWriter, err error) {
	switch {
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}