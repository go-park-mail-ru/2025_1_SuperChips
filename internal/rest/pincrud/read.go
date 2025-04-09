package rest

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

// ReadHandler godoc
// @Summary Get public pin by ID or private pin if user its author
// @Description Returns Pin Data
// @Produce json
// @Param id query int true "requested pin"
// @Success 200 string serverResponse.Data "OK"
// @Failure 400 string serverResponse.Description "invalid query parameter [id]"
// @Failure 403 string serverResponse.Description "access to private pin is forbidden"
// @Failure 404 string serverResponse.Description "no pin with given id"
// @Failure 500 string serverResponse.Description "untracked error: ${error}"
// @Router GET /api/v1/flow
func (app PinCRUDHandler) ReadHandler(w http.ResponseWriter, r *http.Request) {
	pinID, err := parsePinID(r.URL.Query().Get("id"))
	if err != nil {
		rest.HttpErrorToJson(w, "invalid query parameter [id]", http.StatusBadRequest)
		return
	}

	// [TODO] Выяснение, залогинен ли пользователь, через сервис аутентификации.
	var isLogged bool = true
	var userID uint64 = 42

	var data domain.PinData
	if isLogged {
		data, err = app.PinService.GetAnyPin(pinID, userID)
	} else {
		data, err = app.PinService.GetPublicPin(pinID)
	}
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
			msg = "untracked error: " + err.Error()
			status = http.StatusInternalServerError
		}
		rest.HttpErrorToJson(w, msg, status)
		return
	}

	response := rest.ServerResponse{
		Description: "OK",
		Data:        data,
	}
	rest.ServerGenerateJSONResponse(w, response, http.StatusOK)
}
