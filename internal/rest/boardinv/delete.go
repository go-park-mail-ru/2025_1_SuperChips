package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// DeleteInvitation godoc
//	@Summary		Delete invitation link to the board
//	@Description	Delete invitation link to the board for authenticated users
//	@Tags			boards invitations
//	@Produce		json
//	@Security		jwt_auth
//	@Param			board_id	path		int				true	"ID of the board"
//	@Param			link		path		int				true	"Link to the board"
//	@Success		200			{object}	ServerResponse	"Link"
//	@Failure		400			{object}	ServerResponse	"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - access denied"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//	@Router			/api/v1/boards/{board_id}/invitations/{link} [delete]
func (b *BoardInvHandler) DeleteInvitation(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err := strconv.Atoi(boardIDStr)
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	link := r.PathValue("link")
	if link == "" {
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

	err = b.BoardInvService.DeleteInvitation(ctx, boardID, userID, link)
	if err != nil {
		rest.HttpErrorToJson(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := rest.ServerResponse{
		Description: "OK",
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}
