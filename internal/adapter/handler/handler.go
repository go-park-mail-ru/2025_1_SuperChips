package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/usecase"
)

type AppHandler struct {
	Config      configs.Config
	UserService usecase.UserUsecase
	PinService  usecase.PinService
	JWTManager  usecase.JWTManager
}

type serverResponse struct {
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

var ErrBadRequest = fmt.Errorf("bad request")

// HealthCheckHandler godoc
// @Summary Check server status
// @Description Returns server status
// @Produce json
// @Success 200 string serverResponse.Description
// @Router /health [get]
func (app AppHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := serverResponse{
		Description: "server is up",
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

func CorsMiddleware(next http.HandlerFunc, cfg configs.Config, allowedMethods []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if !slices.Contains(allowedMethods, r.Method) {
			httpErrorToJson(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		allowedOrigins := cfg.AllowedOrigins
		if cfg.Environment == "prod" {
			origin := r.Header.Get("Origin")
			if slices.Contains(allowedOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				httpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func serverGenerateJSONResponse(w http.ResponseWriter, body interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		httpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func decodeData(w http.ResponseWriter, body io.ReadCloser, placeholder any) error {
	defer body.Close()

	if err := json.NewDecoder(body).Decode(placeholder); err != nil {
		httpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return err
	}

	return nil
}

// все ответы должны быть json,
// так что это функция для преобразования http ошибок в json
func httpErrorToJson(w http.ResponseWriter, err string, status int) {
	response := serverResponse{
		Description: err,
	}

	serverGenerateJSONResponse(w, response, status)
}
