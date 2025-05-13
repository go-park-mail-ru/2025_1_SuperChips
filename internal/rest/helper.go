package rest

import (
	"io"
	"net/http"

	"github.com/mailru/easyjson"
)

//easyjson:json
type ServerResponse struct {
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

func ServerGenerateJSONResponse(w http.ResponseWriter, body easyjson.Marshaler, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	res, err := easyjson.Marshal(body)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	if _, err := w.Write(res); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func DecodeData(w http.ResponseWriter, body io.ReadCloser, placeholder easyjson.Unmarshaler) error {
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	if err := easyjson.Unmarshal(bodyBytes, placeholder); err != nil {
		if w != nil {
			HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}
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
