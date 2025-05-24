package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// RefuseCoauthoring godoc
//	@Summary		Refuse coauthoring
//	@Description	Refuse coauthoring of the board (user must be coauthor of the board)
//	@Tags			Board sharing [coauthor]
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			board_id	path		int				true	"ID of the board"
//
//	@Success		200			{object}	ServerResponse	"User has stopped being a coauthor"
//	@Failure		400			{object}	ServerResponse	"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse	"Unauthorized"
//	@Failure		403			{object}	ServerResponse	"Forbidden - access denied"
//	@Failure		500			{object}	ServerResponse	"Internal server error"
//
//	@Router			/api/v1/boards/{board_id}/coauthoring [delete]
func (b *BoardShrHandler) RefuseCoauthoring(w http.ResponseWriter, r *http.Request) {
	boardIDStr := r.PathValue("board_id")
	boardID, err:= strconv.Atoi(boardIDStr)
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

	err = b.BoardShrService.RefuseCoauthoring(ctx, boardID, userID)
	if err != nil {
		handleBoardShrError(w, err)
		return
	}

	resp := rest.ServerResponse{
		Description: "OK",
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}