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

// ReassignStarProperty godoc
//
//	@Summary		Reassign star
//	@Description	Reassign star property from pin A with given ID to pin B. Authorization is required. Pins must belong to user.
//	@Tags			Star
//	@Produce		json
//	@Security		jwt_auth
//
//	@Param			flow_id	path		int							true	"initial star pin ID"
//	@Param			flow_id	body		domain.RequestBodyFlowID	true	"new star pin ID"
//
//	@Success		200		{object}	ServerResponse				"Star property has been unset successfully"
//	@Failure		400		{object}	ServerResponse				"Invalid request parameters"
//	@Failure		401		{object}	ServerResponse				"Unauthorized"
//	@Failure		403		{object}	ServerResponse				"Forbidden - access denied / No free star slots"
//	@Failure		404		{object}	ServerResponse				"User doesn't have one of the pins with that ID"
//	@Failure		409		{object}	ServerResponse				"New pin already have star property / Old pin doesn't have star property"
//	@Failure		500		{object}	ServerResponse				"Internal server error"
//
//	@Router			/api/v1/stars/{flow_id} [put]
func (h *StarHandler) ReassignStarProperty(w http.ResponseWriter, r *http.Request) {
	oldPinIDStr := r.PathValue("flow_id")
	oldPinID, err := strconv.Atoi(oldPinIDStr)
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var data domain.RequestBodyFlowID
	if err := rest.DecodeData(w, r.Body, &data); err != nil {
		return
	}
	if data.FlowID == nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var newPinID int = *data.FlowID

	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	userID := claims.UserID

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, h.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(oldPinID > 0 && newPinID > 0 && userID >= 0, "id", "pin id cannot be less or equal to zero or user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	err = h.StarService.ReassignStarProperty(ctx, userID, oldPinID, newPinID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := rest.ServerResponse{
		Description: "OK",
	}

	rest.ServerGenerateJSONResponse(w, resp, http.StatusOK)
}
