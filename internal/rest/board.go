package rest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type BoardService interface {
	CreateBoard(board domain.Board, username string) (int, error)
	DeleteBoard(boardID, userID int) error
	UpdateBoard(boardID, userID int, newName string, isPrivate bool) error 
	AddToBoard(boardID, userID, flowID int) error      // == update board
	DeleteFromBoard(boardID, userID, flowID int) error // == update board
	GetBoard(boardID, userID int, authorized bool) (domain.Board, error)
	GetUserPublicBoards(username string) ([]domain.Board, error)
	GetUserAllBoards(userID int) ([]domain.Board, error)
	GetBoardFlow(boardID, userID, page int, authorized bool) ([]domain.PinData, error)
}

type BoardHandler struct {
	BoardService BoardService
}

type BoardRequest struct {
	FlowID int `json:"flow_id,omitempty"`
}

func (b *BoardHandler) CreateBoardHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if len(username) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var board domain.Board
	board.AuthorID = claims.UserID

	if err := DecodeData(w, r.Body, &board); err != nil {
		return
	}

	id, err := b.BoardService.CreateBoard(board, username)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	type boardId struct {
		BoardID int `json:"board_id"`
	}

	data := boardId{BoardID: id}

	response := ServerResponse{
		Description: "OK",
		Data: data,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

// DELETE api/v1/boards/{board_id}
func (b *BoardHandler) DeleteBoardHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("board_id")
	boardID, err := strconv.Atoi(idString)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	err = b.BoardService.DeleteBoard(boardID, claims.UserID)
	if err != nil {
		handleBoardError(w, err)
		return
	}
	
	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// POST /api/v1/boards/{id}/flow
func (b *BoardHandler) AddToBoardHandler(w http.ResponseWriter, r *http.Request) {
    boardIDStr := r.PathValue("id")
    boardID, err := strconv.Atoi(boardIDStr)
    if err != nil || boardIDStr == "" {
        HttpErrorToJson(w, "Invalid board ID", http.StatusBadRequest)
        return
    }

    claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
    if !ok || claims == nil {
        HttpErrorToJson(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	var request BoardRequest
	if err := DecodeData(w, r.Body, &request); err != nil {
		return
	}

    err = b.BoardService.AddToBoard(boardID, claims.UserID, request.FlowID)
    if err != nil {
		println(err.Error())
        handleBoardError(w, err)
        return
    }

	resp := ServerResponse{
		Description: "OK",
	}

    ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// DELETE /api/v1/boards/{board_id}/flow/{id}
func (b *BoardHandler) DeleteFromBoardHandler(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		HttpErrorToJson(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	flowIDStr := r.PathValue("id")
	flowID, err := strconv.Atoi(flowIDStr)
	if err != nil {
		HttpErrorToJson(w, "Invalid flow ID", http.StatusBadRequest)
		return
	}

	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
    if !ok || claims == nil {
        HttpErrorToJson(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

	if err = b.BoardService.DeleteFromBoard(boardID, claims.UserID, flowID); err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

    ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// PUT /api/v1/boards/{board_id}
func (b *BoardHandler) UpdateBoardHandler(w http.ResponseWriter, r *http.Request) {
    boardIDStr := r.PathValue("board_id")
    boardID, err := strconv.Atoi(boardIDStr)
    if err != nil || boardIDStr == "" {
        HttpErrorToJson(w, "Invalid board ID", http.StatusBadRequest)
        return
    }

    claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
    if !ok || claims == nil {
        HttpErrorToJson(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var updateData struct {
        Name      string `json:"name"`
        IsPrivate bool   `json:"is_private"`
    }

    if err := DecodeData(w, r.Body, &updateData); err != nil {
        return
    }

    err = b.BoardService.UpdateBoard(
        boardID,
        claims.UserID,
        updateData.Name,
        updateData.IsPrivate,
    )
    if err != nil {
        handleBoardError(w, err)
        return
    }

    resp := ServerResponse{
        Description: "OK",
    }

    ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GET /api/v1/boards/{board_id}
func (b *BoardHandler) GetBoardHandler(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	var userID int
	authorized := true

	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		authorized = false
	} else {
		userID = claims.UserID
	}

	board, err := b.BoardService.GetBoard(boardID, userID, authorized)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: board,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GET /api/v1/user/{username}/boards
func (b *BoardHandler) GetUserPublicHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	boards, err := b.BoardService.GetUserPublicBoards(username)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: boards,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GET /api/v1/profile/boards
func (b *BoardHandler) GetUserAllBoardsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	boards, err := b.BoardService.GetUserAllBoards(claims.UserID)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: boards,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GEt /api/v1/boards/{board_id}/flows
func (b *BoardHandler) GetBoardFlowsHandler(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var userID int
	authorized := true

	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		authorized = false
		userID = claims.UserID
	}

	flows, err := b.BoardService.GetBoardFlow(boardID, userID, page, authorized)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
		Data: flows,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func handleBoardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNoBoardName):
		HttpErrorToJson(w, "board name cannot be empty", http.StatusBadRequest)
		return
	case errors.Is(err, domain.ErrBoardAlreadyExists):
		HttpErrorToJson(w, "board already exists", http.StatusConflict)
		return
	case errors.Is(err, domain.ErrConflict):
		HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
	case errors.Is(err, board.ErrForbidden):
		HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	case errors.Is(err, repository.ErrNotFound):
		HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

