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

// DeleteCoauthor godoc
//	@Summary		Remove coauthor
//	@Description	Remove coauthor from the board (user must be author of the board)
//	@Tags			Board sharing [author]
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			board_id	path		int				true	"ID of the board"
//	@Param			name		body		string			true	"Username of coauthor"
//
//	@Success		200			{object}	ServerResponse	"Coauthor has been successfully deleted"
//	@Failure		400			{object}	ServerResponse	"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - access denied"
//	@Failure		404			{object}	ServerResponse	"Username doesn't exist"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//
//	@Router			/api/v1/boards/{board_id}/coauthors [delete]
func (b *BoardShrHandler) DeleteCoauthor(w http.ResponseWriter, r *http.Request) {
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

	var body domain.BodyWithUsername
	if err := rest.DecodeData(w, r.Body, &body); err != nil {
		return
	}
	if body.Name == "" {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	v := validator.New()

	if !v.Check(boardID > 0 && userID >= 0, "id", "board id cannot be less or equal to zero or user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	err = b.BoardShrService.DeleteCoauthor(ctx, boardID, userID, body.Name)
	if err != nil {
		handleBoardShrError(w, err)
		return
	}

	resp := rest.ServerResponse{
		Description: "OK",
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}
