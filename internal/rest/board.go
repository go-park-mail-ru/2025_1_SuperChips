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
	CreateBoard(board domain.Board) error 
	DeleteBoard(boardID, userID int) error
	UpdateBoard(board domain.Board, userID int, newName *string, isPrivate *bool) error 
	AddToBoard(boardID, userID, flowID int) error      // == update board
	DeleteFromBoard(boardID, userID, flowID int) error // == update board
	GetBoard(boardID, userID int) (domain.Board, error)
	GetUserPublicBoards(userID int) ([]domain.Board, error)
	GetUserAllBoards(userID int) ([]domain.Board, error)
}

type BoardHandler struct {
	boardService BoardService
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

	if err := b.boardService.CreateBoard(board); err != nil {
		handleBoardError(w, err)
		return
	}

	response := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

// DELETE api/v1/boards/{id}
func (b *BoardHandler) DeleteBoardHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
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

	err = b.boardService.DeleteBoard(boardID, claims.UserID)
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
	if err := DecodeData(w, r.Body, request); err != nil {
		return
	}

    err = b.boardService.AddToBoard(boardID, claims.UserID, request.FlowID)
    if err != nil {
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

	if err = b.boardService.DeleteFromBoard(boardID, claims.UserID, flowID); err != nil {
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
        Name      *string `json:"name"`
        IsPrivate *bool   `json:"is_private"`
    }

    if err := DecodeData(w, r.Body, &updateData); err != nil {
        return
    }

    if updateData.Name == nil && updateData.IsPrivate == nil {
        HttpErrorToJson(w, "No fields to update", http.StatusBadRequest)
        return
    }

    err = b.boardService.UpdateBoard(
        domain.Board{Id: boardID},
        claims.UserID,
        updateData.Name,
        updateData.IsPrivate,
    )
    if err != nil {
        handleBoardError(w, err)
        return
    }

    resp := ServerResponse{
        Description: "Board updated successfully",
    }

    ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (b *BoardHandler) GetBoardHandler(w http.ResponseWriter, r *http.Request) {
	// todo
	// not implemented yet
	// under construction
}

func handleBoardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNoBoardName):
		HttpErrorToJson(w, "board name cannot be empty", http.StatusBadRequest)
		return
	case errors.Is(err, domain.ErrBoardAlreadyExists):
		HttpErrorToJson(w, "board already exists", http.StatusConflict)
		return
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