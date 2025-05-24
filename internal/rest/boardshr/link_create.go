package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	service "github.com/go-park-mail-ru/2025_1_SuperChips/boardshr"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// CreateInvitation godoc
//	@Summary		Create link
//	@Description	Create invitation link to the board with parameters (person, time limit, usage limit) (user must be author of the board)
//	@Tags			Board sharing [author]
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			board_id	path		int															true	"ID of the board"
//	@Param			names		body		[]string													false	"Usernames for personal invitation"
//	@Param			time_limit	body		time.Time													false	"Time limit for link activity"
//	@Param			usage_limit	body		int															false	"Usage limit"
//
//	@Success		200			{object}	ServerResponse{data=object{link=string}}					"Link has been successfully created"
//	@Success		207			{object}	ServerResponse{data=object{link=string,invalid=[]string}}	"Link has been successfully created for valid names; Invalid usernames are returned"
//	@Failure		400			{object}	ServerResponse												"Invalid request parameters"
//	@Failure		401			{object}	ServerResponse												"Unauthorized"
//	@Failure		404			{object}	ServerResponse												"Link not found"
//	@Failure		403			{object}	ServerResponse												"Forbidden - access denied"
//	@Failure		500			{object}	ServerResponse												"Internal server error"
//
//	@Router			/api/v1/boards/{board_id}/invites [post]
func (b *BoardShrHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
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

	var invitation domain.Invitaion
	if err := rest.DecodeData(w, r.Body, &invitation); err != nil {
		return
	}

	v := validator.New()

	if !v.Check(boardID > 0 && userID >= 0, "id", "board id cannot be less or equal to zero or user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	link, invalidNames, err := b.BoardShrService.CreateInvitation(ctx, boardID, userID, invitation)
	if err != nil && !errors.Is(err, service.ErrNonExistentUsername) {
		handleBoardShrError(w, err)
		return
	}

	// Случай: некоторые имена оказались некорректными.
	description := "OK"
	statusCode := http.StatusOK
	if errors.Is(err, service.ErrNonExistentUsername) {
		description = http.StatusText(http.StatusMultiStatus)
		statusCode = http.StatusMultiStatus
	}

	type DataReturn struct {
		Link         string   `json:"link"`
		InvalidNames []string `json:"invalid,omitempty"`
	}

	resp := rest.ServerResponse{
		Description: description,
		Data:        DataReturn{Link: link, InvalidNames: invalidNames},
	}

	rest.ServerGenerateJSONResponse(w, resp, statusCode)
}
