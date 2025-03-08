package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/feed"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type AppHandler struct {
	Config  configs.Config
	UserStorage user.MapUserStorage
	PinStorage feed.PinSlice
}

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type serverResponse struct {
	Description string      `json:"description,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

var ErrBadRequest = fmt.Errorf("bad request")

func setCookieJWT(w http.ResponseWriter, config configs.Config, email string, userID uint64) error {
    tokenString, err := auth.CreateJWT(config, userID, email)
    if err != nil {
        return err
    }

    http.SetCookie(w, &http.Cookie{
        Name:     auth.AuthToken,
        Value:    tokenString,
        Path:     "/",
        HttpOnly: true,
        Secure:   config.CookieSecure,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Now().Add(config.ExpirationTime),
    })

    return nil
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

		allowedOrigins := []string{"http://localhost:8080", "http://146.185.208.105:8000"}
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
	var authErr user.StatusError

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

func decodeData[T any](w http.ResponseWriter, body io.ReadCloser, placeholder *T) error {
	defer body.Close()

	if err := json.NewDecoder(body).Decode(placeholder); err != nil {
		handleError(w, fmt.Errorf("%w: %s", ErrBadRequest, err.Error()))
		return err
	}

	return nil
}

