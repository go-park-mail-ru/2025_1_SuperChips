package rest

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

// UpdateHandler godoc
// @Summary Update certain pin's fields by ID if user is its author
// @Description Returns JSON with result description
// @Produce JSON
// @Param id body int true "pin ID"
// @Param header body string false "text header"
// @Param description body string false "text description"
// @Param is_private body bool false "privacy setting"
// @Success 200 string serverResponse.Data "OK"
// @Failure 400 string serverResponse.Description "required field is missing [flow_id]"
// @Failure 401 string serverResponse.Description "user is not authorized"
// @Failure 400 string serverResponse.Description "no fields to update"
// @Failure 403 string serverResponse.Description "access to private pin is forbidden"
// @Failure 404 string serverResponse.Description "no pin with given id"
// @Failure 500 string serverResponse.Description "untracked error: ${error}"
// @Router PUT /api/v1/flows
func (app PinCRUDHandler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok {
		rest.HttpErrorToJson(w, "user is not authorized", http.StatusUnauthorized)
		return
	}
	userID := uint64(claims.UserID)

	data := domain.PinDataUpdate{}
	if err := rest.DecodeData(w, r.Body, &data); err != nil {
		return
	}
	if data.FlowID == nil {
		rest.HttpErrorToJson(w, "required field is missing [flow_id]", http.StatusBadRequest)
		return
	}

	err := app.PinService.UpdatePin(data, userID)
	if errors.Is(err, pincrud.ErrForbidden) {
		rest.HttpErrorToJson(w, "access to private pin is forbidden", http.StatusForbidden)
		return
	}
	if errors.Is(err, pincrud.ErrPinNotFound) {
		rest.HttpErrorToJson(w, "no pin with given id", http.StatusNotFound)
		return
	}
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := rest.ServerResponse{
		Description: "OK",
	}
	rest.ServerGenerateJSONResponse(w, response, http.StatusOK)
}
