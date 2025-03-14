package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/feed"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
	statusError "github.com/go-park-mail-ru/2025_1_SuperChips/internal/error"
)

type AppHandler struct {
	Config  configs.Config
	UserStorage user.MapUserStorage
	PinStorage feed.PinStorage
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
			handleHttpError(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		allowedOrigins := []string{"http://localhost:8080", "http://146.185.208.105:8000", "http://localhost:8000"}
        if cfg.Environment == "prod" {
            origin := r.Header.Get("Origin")
            if slices.Contains(allowedOrigins, origin) {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            } else {
				handleHttpError(w, "Forbidden", http.StatusForbidden)
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

func handleHttpError(w http.ResponseWriter, errorDesc string, statusCode int) {
	errorResp := serverResponse{
		Description: errorDesc,
	}

	serverGenerateJSONResponse(w, errorResp, statusCode)
}

func handleError(w http.ResponseWriter, err error) {
	var authErr statusError.StatusError

	errorResp := serverResponse{
		Description: http.StatusText(http.StatusInternalServerError),
	}

	if errors.As(err, &authErr) {
		errorResp.Description = http.StatusText(authErr.StatusCode())
		serverGenerateJSONResponse(w, errorResp, authErr.StatusCode())
		return
	}

	if errors.Is(err, http.ErrNoCookie) {
		errorResp.Description = http.StatusText(http.StatusForbidden)
		serverGenerateJSONResponse(w, errorResp, http.StatusForbidden)
		return
	}

	if errors.Is(err, ErrBadRequest) {
		errorResp.Description = http.StatusText(http.StatusBadRequest)
		serverGenerateJSONResponse(w, errorResp, http.StatusBadRequest)
		return
	}

	serverGenerateJSONResponse(w, errorResp, http.StatusInternalServerError)
}

func serverGenerateJSONResponse(w http.ResponseWriter, body interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		handleError(w, err)
	}
}

func decodeData(w http.ResponseWriter, body io.ReadCloser, placeholder any) error {
	defer body.Close()

	if err := json.NewDecoder(body).Decode(placeholder); err != nil {
		handleError(w, fmt.Errorf("%w: %s", ErrBadRequest, err.Error()))
		return err
	}

	return nil
}

