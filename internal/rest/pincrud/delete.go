package rest

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

// DeleteHandler godoc
// @Summary Delete pin by ID if user is its author
// @Description Returns JSON with result description
// @Produce JSON
// @Param id query int true "pin to delete"
// @Success 200 string serverResponse.Data "OK"
// @Failure 400 string serverResponse.Description "invalid query parameter [id]"
// @Failure 403 string serverResponse.Description "access to private pin is forbidden"
// @Failure 404 string serverResponse.Description "no pin with given id"
// @Failure 500 string serverResponse.Description "untracked error: ${error}"
// @Router DELETE /api/v1/flows
func (app PinCRUDHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok {
		rest.HttpErrorToJson(w, "user is not authorized", http.StatusUnauthorized)
		return
	}
	userID := uint64(claims.UserID)

	pinID, err := parsePinID(r.URL.Query().Get("id"))
	if err != nil {
		rest.HttpErrorToJson(w, "invalid query parameter [id]", http.StatusBadRequest)
		return
	}

	err = app.PinService.DeletePinByID(pinID, userID)
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
