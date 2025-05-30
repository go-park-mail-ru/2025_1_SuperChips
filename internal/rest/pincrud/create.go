package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/pincrud"
)

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/bmp":  true,
	"image/tiff": true,
	"image/gif": true,
}

const (
	maxPinSize = 1024 * 1024 * 10 // 10 mb
)

// CreateHandler godoc
//	@Summary		Create pin if user if user is authorized
//	@Description	Returns JSON with result description
//	@Produce		json
//	@Param			image		formData	file						true	"pin image"
//	@Param			header		formData	string						false	"text header"
//	@Param			description	formData	string						false	"text description"
//	@Param			is_private	formData	bool						false	"privacy setting"
//	@Success		201			string		serverResponse.Data			"OK"
//	@Failure		400			string		serverResponse.Description	"failed to parse the request body"
//	@Failure		400			string		serverResponse.Description	"image not present in the request body"
//	@Failure		400			string		serverResponse.Description	"failed to parse the form-data field [is_private]"
//	@Failure		400			string		serverResponse.Description	"invalid image extension"
//	@Failure		401			string		serverResponse.Description	"user is not authorized"
//	@Failure		500			string		serverResponse.Description	"untracked error: ${error}"
//	@Router			/api/v1/flows [post]
func (app PinCRUDHandler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok {
		rest.HttpErrorToJson(w, "user is not authorized", http.StatusUnauthorized)
		return
	}
	userID := uint64(claims.UserID)

	err := r.ParseMultipartForm(10 << 20) // 10 Мбайт.
	if err != nil {
		rest.HttpErrorToJson(w, "failed to parse the request body", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		rest.HttpErrorToJson(w, "image not present in the request body", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if handler.Size > maxPinSize {
		rest.HttpErrorToJson(w, "image is too big", http.StatusRequestEntityTooLarge)
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	detected := http.DetectContentType(buffer)
	contentType := handler.Header.Get("Content-Type")

	if !strings.HasPrefix(detected, strings.Split(contentType, ";")[0]) {
		rest.HttpErrorToJson(w, "image extension and type are mismatched", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if _, ok := allowedTypes[contentType]; !ok {
		rest.HttpErrorToJson(w, "image type is not allowed", http.StatusBadRequest)
		return
	}

	filename := filepath.Base(handler.Filename)
    ext := filepath.Ext(filename)
    if ext == "" {
        rest.HttpErrorToJson(w, "invalid file extension", http.StatusBadRequest)
        return
    }

	data := domain.PinDataCreate{
		Header:      "",
		Description: "",
		IsPrivate:   false,
	}

	if r.PostFormValue("header") != "" {
		data.Header = r.PostFormValue("header")
	}
	if r.PostFormValue("description") != "" {
		data.Description = r.PostFormValue("description")
	}
	if r.PostFormValue("is_private") != "" {
		boolValue, err := strconv.ParseBool(r.PostFormValue("is_private"))
		if err != nil {
			rest.HttpErrorToJson(w, "failed to parse the form-data field [is_private]", http.StatusBadRequest)
			return
		}
		data.IsPrivate = boolValue
	}

	pinID, name, err := app.PinService.CreatePin(r.Context(), data, file, handler, contentType, userID)
	if errors.Is(err, pincrud.ErrInvalidImageExt) {
		rest.HttpErrorToJson(w, "invalid image extension", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("create flow err: %v", err)
		rest.HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := sendRequestToCV(name); err != nil {
		log.Printf("sending request to cv error: %v", err)
		// no return
	}

	type DataReturn struct {
		FlowID uint64 `json:"flow_id"`
	}

	response := rest.ServerResponse{
		Description: "OK",
		Data:        DataReturn{FlowID: pinID},
	}
	rest.ServerGenerateJSONResponse(w, response, http.StatusCreated)
}

func sendRequestToCV(filename string) error {
	type Filename struct {
		Filename string `json:"filename"`
	}

	pinFilename := Filename{
		Filename: filename,
	}

	filenameBytes, err := json.Marshal(pinFilename)
	if err != nil {
		log.Printf("Error making request to cv: %v", err)
		return err
	} 

	req, err := http.NewRequest("POST", "http://cv:8050/classify", bytes.NewBuffer(filenameBytes))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Unexpected response status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}