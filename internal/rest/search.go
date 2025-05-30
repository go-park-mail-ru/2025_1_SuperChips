package rest

import (
	"context"
	"log"
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

// SearchPins godoc
//	@Summary		Searches for pins
//	@Description	Returns a pageSized number of pins searched for
//	@Produce		json
//	@Param			page	path	int							true	"requested page"		example("?page=3")
//	@Param			size	path	int							true	"requested page size"	example("?size=15")
//	@Param			query	path	string						true	"search query"			example("?query=kittens")
//	@Success		200		string	serverResponse.Data			"OK"
//	@Failure		400		string	serverResponse.Description	"bad request"
//	@Failure		404		string	serverResponse.Description	"page not found"
//	@Failure		500		string	serverResponse.Description	"internal server error"
//	@Router			/api/v1/search/pins [get]
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
		log.Printf("search pin error: %v", err)
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

// SearchBoards godoc
//	@Summary		Searches for boards
//	@Description	Returns a pageSized number of boards searched for
//	@Produce		json
//	@Param			page	path	int							true	"requested page"		example("?page=3")
//	@Param			size	path	int							true	"requested page size"	example("?size=15")
//	@Param			query	path	string						true	"search query"			example("?query=kittens")
//	@Success		200		string	serverResponse.Data			"OK"
//	@Failure		400		string	serverResponse.Description	"bad request"
//	@Failure		404		string	serverResponse.Description	"page not found"
//	@Failure		500		string	serverResponse.Description	"internal server error"
//	@Router			/api/v1/search/boards [get]
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
		log.Printf("search boards error: %v", err)
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

// SearchUsers godoc
//	@Summary		Searches for users
//	@Description	Returns a pageSized number of users searched for
//	@Produce		json
//	@Param			page	path	int							true	"requested page"		example("?page=3")
//	@Param			size	path	int							true	"requested page size"	example("?size=15")
//	@Param			query	path	string						true	"search query"			example("?query=kittens")
//	@Success		200		string	serverResponse.Data			"OK"
//	@Failure		400		string	serverResponse.Description	"bad request"
//	@Failure		404		string	serverResponse.Description	"page not found"
//	@Failure		500		string	serverResponse.Description	"internal server error"
//	@Router			/api/v1/search/users [get]
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
		log.Printf("search users error: %v", err)
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