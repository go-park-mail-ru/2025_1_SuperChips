package rest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

// CreateHandler godoc
// @Summary Create pin if user if user is authorized
// @Description Returns JSON with result description
// @Produce JSON
// @Param image formData file true "pin image"
// @Param header formData string false "text header"
// @Param description formData string false "text description"
// @Param is_private formData bool false "privacy setting"
// @Success 201 string serverResponse.Data "OK"
// @Failure 400 string serverResponse.Description "failed to parse the request body"
// @Failure 400 string serverResponse.Description "image not present in the request body"
// @Failure 400 string serverResponse.Description "failed to parse the form-data field [is_private]"
// @Failure 400 string serverResponse.Description "invalid image extension"
// @Failure 401 string serverResponse.Description "user is not authorized"
// @Failure 500 string serverResponse.Description "untracked error: ${error}"
// @Router POST /api/v1/flow
func (app PinCRUDHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	// [TODO] Выяснение, залогинен ли пользователь, через сервис аутентификации.
	var isLogged bool = true
	var userID uint64 = 42
	if !isLogged {
		rest.HttpErrorToJson(w, "user is not authorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 Мбайт.
	if err != nil {
		rest.HttpErrorToJson(w, "failed to parse the request body", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		rest.HttpErrorToJson(w, "image not present in the request body", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data := domain.PinDataCreate{
		Header:      "",
		Description: "",
		IsPrivate:   true,
	}
	if r.PostFormValue("header") != "" {
		data.Header = r.PostFormValue("header")
	}
	if r.PostFormValue("description") != "" {
		data.Description = r.PostFormValue("description")
	}
	if r.PostFormValue("is_private") != "" {
		boolValue, err := strconv.ParseBool(r.FormValue("is_private"))
		if err != nil {
			rest.HttpErrorToJson(w, "failed to parse the form-data field [is_private]", http.StatusBadRequest)
			return
		}
		data.IsPrivate = boolValue
	}

	err = app.PinService.CreatePin(data, file, header, userID)
	if err != nil {
		var msg string
		var status int
		switch {
		case errors.Is(err, pincrud.ErrInvalidImageExt):
			msg = "invalid image extension"
			status = http.StatusBadRequest
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
	rest.ServerGenerateJSONResponse(w, response, http.StatusCreated)
}
