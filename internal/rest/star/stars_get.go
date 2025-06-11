package rest

import (
	"context"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/validator"
)

// SetStarProperty godoc
//
//	@Summary		Get stars
//	@Description	Get star pins of user. Authorization is required.
//	@Tags			Star
//	@Produce		json
//	@Security		jwt_auth
//
//	@Success		200	{object}	ServerResponse{data=object{total_slots_count=int,pins=[]domain.PinData}}	"Star pins have been successfully fetched"
//	@Failure		401	{object}	ServerResponse																"Unauthorized"
//	@Failure		403	{object}	ServerResponse																"Forbidden - access denied"
//	@Failure		500	{object}	ServerResponse																"Internal server error"
//
//	@Router			/api/v1/stars [get]
func (h *StarHandler) GetStarPins(w http.ResponseWriter, r *http.Request) {
	claims, _ := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	userID := claims.UserID

	ctx, cancel := context.WithTimeout(context.Background(), h.ContextDeadline)
	defer cancel()

	v := validator.New()

	if !v.Check(userID >= 0, "id", "user id cannot be less than zero") {
		rest.HttpErrorToJson(w, v.GetError("id").Error(), http.StatusBadRequest)
		return
	}

	totalSlotsCount, starPins, err := h.StarService.GetStarPins(ctx, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	domain.EscapeFlows(starPins)

	type DataReturn struct {
		TotalSlotsCount int              `json:"total_slots_count"`
		Pins            []domain.PinData `json:"pins"`
	}
	response := rest.ServerResponse{
		Description: "OK",
		Data:        DataReturn{TotalSlotsCount: totalSlotsCount, Pins: starPins},
	}

	rest.ServerGenerateJSONResponse(w, response, http.StatusOK)
}
