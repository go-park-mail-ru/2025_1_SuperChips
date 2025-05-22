package rest

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/board"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	repository "github.com/go-park-mail-ru/2025_1_SuperChips/internal/repository/pg"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

type BoardService interface {
	CreateBoard(ctx context.Context, board domain.Board, username string, userID int) (int, error)          // создание доски
	DeleteBoard(ctx context.Context, boardID, userID int) error                                             // удаление доски
	UpdateBoard(ctx context.Context, boardID, userID int, newName string, isPrivate bool) error             // обновление доски
	AddToBoard(ctx context.Context, boardID, userID, flowID int) error                                      // добавить пин в доску
	DeleteFromBoard(ctx context.Context, boardID, userID, flowID int) error                                 // удалить пин из доски
	GetBoard(ctx context.Context, boardID, userID int, authorized bool) (domain.Board, error)               // получить доску
	GetUserPublicBoards(ctx context.Context, username string) ([]domain.Board, error)                       // получить публичные доски пользователя
	GetUserAllBoards(ctx context.Context, userID int) ([]domain.Board, error)                               // получить все доски пользователя
	GetBoardFlow(ctx context.Context, boardID, userID, page, pageSize int, authorized bool) ([]domain.PinData, error) // получить пины доски
}

type BoardHandler struct {
	BoardService    BoardService
	ContextDeadline time.Duration
}

// CreateBoard godoc
//	@Summary		Create a new board
//	@Description	Creates a new board for the specified user
//	@Tags			boards
//	@Accept			json
//	@Produce		json
//	@Security		jwt_auth
//	@Param			username	path		string			true	"Username of the board owner"
//	@Param			board		body		domain.Board	true	"Board details"
//	@Success		200			{object}	ServerResponse	"Board created successfully"
//	@Failure		400			{object}	ServerResponse	"Invalid request data"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		409			{object}	ServerResponse	"Board already exists"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{username} [post]
func (b *BoardHandler) CreateBoard(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if len(username) == 0 {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var board domain.Board
	board.AuthorID = claims.UserID

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	if err := DecodeData(w, r.Body, &board); err != nil {
		return
	}

	v := validator.New()

	if !v.Check(board.Name != "", "name", "cannot be empty") {
		HttpErrorToJson(w, v.GetError("name").Error(), http.StatusBadRequest)
		return
	}

	if !v.Check(len(board.Name) < 64, "name", "cannot be longer 64") {
		HttpErrorToJson(w, v.GetError("name").Error(), http.StatusBadRequest)
		return
	}


	id, err := b.BoardService.CreateBoard(ctx, board, username, claims.UserID)
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
		Data:        data,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

// DeleteBoard godoc
//	@Summary		Delete a board
//	@Description	Deletes a board by ID for authenticated user
//	@Tags			boards
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int				true	"ID of the board to delete"
//	@Success		200			{object}	ServerResponse	"Board deleted successfully"
//	@Failure		400			{object}	ServerResponse	"Invalid board ID"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - not board owner"
//	@Failure		404			{object}	ServerResponse	"Board not found"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{board_id} [delete]
func (b *BoardHandler) DeleteBoard(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("board_id")
	boardID, err := strconv.Atoi(idString)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(claims.UserID > 0 && boardID > 0, "id", "cannot be less or equal to zero") {
		HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	err = b.BoardService.DeleteBoard(ctx, boardID, claims.UserID)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// AddToBoard godoc
//	@Summary		Add flow to board
//	@Description	Adds a flow to a board for authenticated user
//	@Tags			boards
//	@Accept			json
//	@Produce		json
//	@Security		jwt_auth
//	@Param			id	path		int				true	"Board ID"
// //	@Param			flow	body		BoardRequest	true	"Flow ID to add"
//	@Success		200	{object}	ServerResponse	"Flow added successfully"
//	@Failure		400	{object}	ServerResponse	"Invalid request data"
//	@Failure		401	{object}	ServerResponse	"Unauthorized"
//	@Failure		403	{object}	ServerResponse	"Forbidden - not board owner"
//	@Failure		404	{object}	ServerResponse	"Board or flow not found"
//	@Failure		500	{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{id}/flows [post]
func (b *BoardHandler) AddToBoard(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil || boardIDStr == "" {
		HttpErrorToJson(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	var request domain.BoardRequest
	if err := DecodeData(w, r.Body, &request); err != nil {
		return
	}

	v := validator.New()

	if !v.Check(request.FlowID > 0 && boardID > 0 && claims.UserID > 0, "id", "cannot be less or equal to zero") {
		HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	err = b.BoardService.AddToBoard(ctx, boardID, claims.UserID, request.FlowID)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// DeleteFromBoard godoc
//	@Summary		Remove flow from board
//	@Description	Removes a flow from a board for authenticated user
//	@Tags			boards
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int				true	"Board ID"
//	@Param			id			path		int				true	"Flow ID to remove"
//	@Success		200			{object}	ServerResponse	"Flow removed successfully"
//	@Failure		400			{object}	ServerResponse	"Invalid request data"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - not board owner"
//	@Failure		404			{object}	ServerResponse	"Board or flow not found"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{board_id}/flows/{id} [delete]
func (b *BoardHandler) DeleteFromBoard(w http.ResponseWriter, r *http.Request) {
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

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(flowID > 0 && boardID > 0 && claims.UserID > 0, "id", "cannot be less or equal to zero") {
		HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	if err = b.BoardService.DeleteFromBoard(ctx, boardID, claims.UserID, flowID); err != nil {
		handleBoardError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// UpdateBoard godoc
//	@Summary		Update board details
//	@Description	Updates board name and privacy settings
//	@Tags			boards
//	@Accept			json
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int				true	"Board ID to update"
//	@Param			updateData	body		object			true	"update data: new name and is_private"
//	@Success		200			{object}	ServerResponse	"Board updated successfully"
//	@Failure		400			{object}	ServerResponse	"Invalid request data"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - not board owner"
//	@Failure		404			{object}	ServerResponse	"Board not found"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{board_id} [put]
func (b *BoardHandler) UpdateBoard(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil || boardIDStr == "" {
		HttpErrorToJson(w, "Invalid board ID", http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)

	var updateData domain.UpdateData

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	if err := DecodeData(w, r.Body, &updateData); err != nil {
		return
	}

	v := validator.New()

	if !v.Check(boardID > 0 && claims.UserID > 0, "id", "cannot be less or equal to zero") {
		HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	if !v.Check(updateData.Name != "", "name", "cannot be empty") {
		HttpErrorToJson(w, v.GetError("name").Error(), http.StatusBadRequest)
		return
	}

	err = b.BoardService.UpdateBoard(
		ctx,
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

// GetBoard godoc
//	@Summary		Get board details
//	@Description	Retrieves board information with access control
//	@Tags			boards
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int									true	"Board ID to retrieve"
//	@Success		200			{object}	ServerResponse{data=domain.Board}	"Board details"
//	@Failure		400			{object}	ServerResponse						"Invalid board ID"
//	@Failure		401			{object}	ServerResponse						"Unauthorized"
//	@Failure		403			{object}	ServerResponse						"Forbidden - private board"
//	@Failure		404			{object}	ServerResponse						"Board not found"
//	@Failure		500			{object}	ServerResponse						"Internal server error"
//	@Router			/api/v1/boards/{board_id} [get]
func (b *BoardHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(boardID > 0 && userID >= 0, "id", "board id cannot be less or equal to zero or user id cannot be less than zero") {
		HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	board, err := b.BoardService.GetBoard(ctx, boardID, userID, authorized)
	if err != nil {
		log.Printf("get board err: %v", err)
		handleBoardError(w, err)
		return
	}

	board.Escape()

	resp := ServerResponse{
		Description: "OK",
		Data:        board,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GetUserPublic godoc
//	@Summary		Get user's public boards
//	@Description	Retrieves public boards for a specific user
//	@Tags			boards
//	@Produce		json
//	@Param			username	path		string								true	"Username to retrieve public boards for"
//	@Success		200			{object}	ServerResponse{data=[]domain.Board}	"Public boards list"
//	@Failure		400			{object}	ServerResponse						"Invalid username"
//	@Failure		404			{object}	ServerResponse						"User not found"
//	@Failure		500			{object}	ServerResponse						"Internal server error"
//	@Router			/api/v1/user/{username}/boards [get]
func (b *BoardHandler) GetUserPublic(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	boards, err := b.BoardService.GetUserPublicBoards(ctx, username)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	domain.EscapeBoards(boards)

	resp := ServerResponse{
		Description: "OK",
		Data:        boards,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GetUserAllBoards godoc
//	@Summary		Get all user boards
//	@Description	Retrieves all boards (public and private) for authenticated user
//	@Tags			boards
//	@Produce		json
//	@Security		jwt_auth
//	@Success		200	{object}	ServerResponse{data=[]domain.Board}	"User's boards list"
//	@Failure		401	{object}	ServerResponse						"Unauthorized"
//	@Failure		500	{object}	ServerResponse						"Internal server error"
//	@Router			/api/v1/profile/boards [get]
func (b *BoardHandler) GetUserAllBoards(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	boards, err := b.BoardService.GetUserAllBoards(ctx, claims.UserID)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	domain.EscapeBoards(boards)

	resp := ServerResponse{
		Description: "OK",
		Data:        boards,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

// GetBoardFlows godoc
//	@Summary		Get board flows with pagination
//	@Description	Retrieves flows in a board with pagination for authenticated users
//	@Tags			boards
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int										true	"ID of the board to retrieve flows from"
//	@Param			page		query		int										true	"Page number (0-based index)"
//	@Param			size		query		int										true	"Number of items per page"
//	@Success		200			{object}	ServerResponse{data=[]domain.PinData}	"List of flows in the board"
//	@Failure		400			{object}	ServerResponse							"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse							"Unauthorized"
//	@Failure		403			{object}	ServerResponse							"Forbidden - access denied"
//	@Failure		404			{object}	ServerResponse							"Board not found"
//	@Failure		500			{object}	ServerResponse							"Internal server error"
//	@Router			/api/v1/boards/{board_id}/flows [get]
func (b *BoardHandler) GetBoardFlows(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 0
	}

	pageSizeStr := r.URL.Query().Get("size")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(boardID > 0 && userID >= 0 || page > 0, "id and page", "cannot be less than zero") {
		HttpErrorToJson(w, v.GetError("id and page").Error(), http.StatusBadRequest)
		return
	}

	if !v.Check(pageSize >= 1 && pageSize <= 30, "page size", "cannot be less than one and more than 30") {
		HttpErrorToJson(w, v.GetError("page size").Error(), http.StatusBadRequest)
		return
	}

	flows, err := b.BoardService.GetBoardFlow(ctx, boardID, userID, page, pageSize, authorized)
	if err != nil {
		handleBoardError(w, err)
		return
	}

	domain.EscapeFlows(flows)

	resp := ServerResponse{
		Description: "OK",
		Data:        flows,
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
