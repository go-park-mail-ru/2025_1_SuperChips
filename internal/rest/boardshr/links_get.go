package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// GetInvitationLinks godoc
//	@Summary		Get links
//	@Description	Get invitation links to the board with ID with parameters (user must be author of the board)
//	@Tags			Board sharing [author]
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			board_id	path		int														true	"ID of the board"
//
//	@Success		200			{object}	ServerResponse{data=object{links=[]domain.LinkParams}}	"Link list has been successfully fetched"
//	@Failure		400			{object}	ServerResponse											"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse											"Unauthorized"
//	@Failure		403			{object}	ServerResponse											"Forbidden - access denied"
//	@Failure		404			{object}	ServerResponse											"Board or links not found"
//	@Failure		500			{object}	ServerResponse											"Internal server error"
//
//	@Router			/api/v1/boards/{board_id}/invites [get]
func (b *BoardShrHandler) GetInvitationLinks(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	userID := claims.UserID

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(boardID > 0 && userID >= 0, "id", "board id cannot be less or equal to zero or user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	links, err := b.BoardShrService.GetInvitationLinks(ctx, boardID, userID)
	if err != nil {
		handleBoardShrError(w, err)
		return
	}
	
	// Сценарий: ссылок нет.
	if len(links) == 0 {
		resp := rest.ServerResponse{
			Description: http.StatusText(http.StatusNotFound),
		}
		rest.ServerGenerateJSONResponse(w, resp, http.StatusNotFound)
		return
	}

	type DataReturn struct {
		Links []domain.LinkParams `json:"links"`
	}

	resp := rest.ServerResponse{
		Description: "OK",
		Data:        DataReturn{Links: links},
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}
