package rest

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
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
// @Router DELETE /api/v1/flow
func (app PinCRUDHandler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	pinID, err := parsePinID(r.URL.Query().Get("id"))
	if err != nil {
		rest.HttpErrorToJson(w, "invalid query parameter [id]", http.StatusBadRequest)
		return
	}

	// [TODO] Выяснение, залогинен ли пользователь, через сервис аутентификации.
	var isLogged bool = true
	var userID uint64 = 42
	if !isLogged {
		rest.HttpErrorToJson(w, "access to private pin is forbidden", http.StatusForbidden)
		return
	}

	err = app.PinService.DeletePinByID(pinID, userID)
	if err != nil {
		var msg string
		var status int
		switch {
		case errors.Is(err, pincrud.ErrForbidden):
			msg = "access to private pin is forbidden"
			status = http.StatusForbidden
		case errors.Is(err, pincrud.ErrPinNotFound):
			msg = "no pin with given id"
			status = http.StatusNotFound
		default:
			msg = err.Error()
			status = http.StatusInternalServerError
		}
		rest.HttpErrorToJson(w, msg, status)
		return
	}

	response := rest.ServerResponse{
		Description: "OK",
	}
	rest.ServerGenerateJSONResponse(w, response, http.StatusOK)
}
