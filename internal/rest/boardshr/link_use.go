package rest

import (
	"context"
	"errors"
	"net/http"

	service "github.com/go-park-mail-ru/2025_1_SuperChips/boardshr"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// UseInvitationLink godoc
//
//	@Summary		Join via link
//	@Description	Join the board via invitation link as co-author; link mustn't be expired and, if link is private, user must be in group
//	@Tags			Board sharing [coauthor]
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			link	path		string										true	"Link"
//
//	@Success		200		{object}	ServerResponse{data=object{board_id=int}}	"User has successfully become a coauthor of the board"
//	@Failure		400		{object}	ServerResponse								"Invalid request parameters"
//	@Failure		401		{object}	ServerResponse								"Unauthorized"
//	@Failure		403		{object}	ServerResponse								"Forbidden - access denied"
//	@Failure		404		{object}	ServerResponse								"Link not found"
//	@Failure		409		{object}	ServerResponse{data=object{board_id=int}}	"User is already coauthor"
//	@Failure		410		{object}	ServerResponse								"Link's time or usage limit has expired"
//	@Failure		500		{object}	ServerResponse								"Internal server error"
//
//	@Router			/api/v1/join/{link} [post]
func (b *BoardShrHandler) UseInvitationLink(w http.ResponseWriter, r *http.Request) {
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

	if !v.Check(userID >= 0, "id", "user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	boardID, err := b.BoardShrService.UseInvitationLink(ctx, userID, link)
	if err != nil && !errors.Is(err, service.ErrAlreadyEditor) {
		handleBoardShrError(w, err)
		return
	}

	// Сценарий: пользователь успешно стал соавтором доски.
	status, description := http.StatusOK, "OK"

	// Сценарий: пользователь уже является соавтором доски.
	if errors.Is(err, service.ErrAlreadyEditor) {
		status, description = http.StatusConflict, http.StatusText(http.StatusConflict)
	}

	type DataReturn struct {
		BoardID int `json:"board_id"`
	}

	resp := rest.ServerResponse{
		Description: description,
		Data:        DataReturn{BoardID: boardID},
	}

	rest.ServerGenerateJSONResponse(w, resp, status)
}
