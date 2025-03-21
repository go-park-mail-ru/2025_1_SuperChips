package rest

import (
	"encoding/json"
	"io"
	"net/http"
)

type ServerResponse struct {
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func ServerGenerateJSONResponse(w http.ResponseWriter, body interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func DecodeData(w http.ResponseWriter, body io.ReadCloser, placeholder any) error {
	defer body.Close()

	if err := json.NewDecoder(body).Decode(placeholder); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return err
	}

	return nil
}

// все ответы должны быть json,
// так что это функция для преобразования http ошибок в json
func HttpErrorToJson(w http.ResponseWriter, err string, status int) {
	response := ServerResponse{
		Description: err,
	}

	ServerGenerateJSONResponse(w, response, status)
}
