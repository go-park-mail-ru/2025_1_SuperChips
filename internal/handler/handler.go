package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type AppHandler struct {
	Config  configs.Config
	Storage user.MapUserStorage
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

func handleError(w http.ResponseWriter, err error) {
	var authErr user.StatusError
	if errors.As(err, &authErr) {
		http.Error(w, http.StatusText(authErr.StatusCode()), authErr.StatusCode())
		return
	}

	if errors.Is(err, http.ErrNoCookie) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	if errors.Is(err, ErrBadRequest) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

func (app AppHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("server is up"))
	if err != nil {
		handleError(w, err)
		return
	}
}

func (app AppHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := loginData{}
	if err := decodeData(w, r.Body, &data); err != nil {
		return
	}

	errorResp := serverResponse{
		Description: "OK",
	}

	if err := app.Storage.LoginUser(data.Email, data.Password); err != nil {
		errorResp.Description = "invalid credentials"
		serverGenerateJSONResponse(w, errorResp, http.StatusForbidden)
		return
	}

	id := app.Storage.GetUserId(data.Email)

	if err := auth.SetCookieJWT(w, app.Config, data.Email, id); err != nil {
		handleError(w, err)
		return
	}

	serverGenerateJSONResponse(w, errorResp, http.StatusOK)
}

func (app AppHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := decodeData(w, r.Body, &userData); err != nil {
		return
	}

	errorResp := serverResponse{
		Description: "OK",
	}

	statusCode := http.StatusOK

	if err := app.Storage.AddUser(userData); err != nil {
		switch {
		case errors.Is(err, user.ErrConflict):
			statusCode = http.StatusConflict
			errorResp.Description = err.Error()
		case errors.Is(err, user.ErrValidation):
			statusCode = http.StatusBadRequest
			errorResp.Description = err.Error()
		default:
			handleError(w, err)
			return
		}
	}

	serverGenerateJSONResponse(w, errorResp, statusCode)
}

func (app AppHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	expiredCookie := &http.Cookie{
		Name:     auth.AuthToken,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-time.Hour * 24 * 365),
	}

	http.SetCookie(w, expiredCookie)

	w.WriteHeader(http.StatusOK)
}

func (app AppHandler) UserDataHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		handleError(w, err)
		return
	}

	claims, err := auth.ParseJWTToken(token.Value, app.Config)
	if err != nil {
		handleError(w, err)
		return
	}

	userData, err := app.Storage.GetUserPublicInfo(claims.Email)
	if err != nil {
		handleError(w, err)
		return
	}
	response := serverResponse{
		Data: &userData,
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

