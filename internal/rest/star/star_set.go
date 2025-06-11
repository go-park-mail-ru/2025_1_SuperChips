package rest

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// SetStarProperty godoc
//
//	@Summary		Set star
//	@Description	Set star property for pin with given ID. Authorization is required. Pin must belong to user.
//	@Tags			Star
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			flow_id	path		int				true	"pin ID"
//
//	@Success		200		{object}	ServerResponse	"Star property has been set successfully"
//	@Failure		400		{object}	ServerResponse	"Invalid request parameters"
//	@Failure		401		{object}	ServerResponse	"Unauthorized"
//	@Failure		403		{object}	ServerResponse	"Forbidden - access denied / No free star slots"
//	@Failure		404		{object}	ServerResponse	"User doesn't have pin with that ID"
//	@Failure		409		{object}	ServerResponse	"Star property was already set"
//	@Failure		500		{object}	ServerResponse	"Internal server error"
//
//	@Router			/api/v1/stars/{flow_id} [post]
func (h *StarHandler) SetStarProperty(w http.ResponseWriter, r *http.Request) {
	pinIDStr := r.PathValue("flow_id")
	pinID, err := strconv.Atoi(pinIDStr)
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	userID := claims.UserID

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, h.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(pinID > 0 && userID >= 0, "id", "pin id cannot be less or equal to zero or user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	err = h.StarService.SetStarProperty(ctx, userID, pinID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := rest.ServerResponse{
		Description: "OK",
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}
