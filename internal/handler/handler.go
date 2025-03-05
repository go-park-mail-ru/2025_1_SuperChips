package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type serverResponse struct {
	Error      string           `json:"error,omitempty"`
	PublicData *user.PublicUser `json:",omitempty"`
}

func handleError(w http.ResponseWriter, err error) {
	switch err {
	case user.ErrEmailAlreadyTaken, user.ErrUsernameAlreadyTaken:
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
	case user.ErrInvalidBirthday, user.ErrInvalidCredentials, user.ErrInvalidEmail, user.ErrInvalidUsername, user.ErrNoPassword, user.ErrPasswordTooLong:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	case http.ErrNoCookie, auth.ErrorExpiredToken:
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	case user.ErrUserNotFound:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	default:
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func serverGenerateJSONResponse(w http.ResponseWriter, body interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		handleError(w, err)
	}
}

// здесь кмк без дженерика тяжеловато будет, так как *interface{} у меня не сработал
func decodeData[T any](w http.ResponseWriter, body io.ReadCloser, placeholder *T) error {
	if err := json.NewDecoder(body).Decode(placeholder); err != nil {
		handleError(w, err)
		return err
	}

	return nil
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("server is up"))
	if err != nil {
		handleError(w, err)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := loginData{}
	if err := decodeData(w, r.Body, &data); err != nil {
		return
	}

	errorResp := serverResponse{
		Error: "OK",
	}

	if err := user.LoginUser(data.Email, data.Password); err != nil {
		errorResp.Error = err.Error()
		serverGenerateJSONResponse(w, errorResp, http.StatusForbidden)
		return
	}

	if err := auth.SetCookieJWT(w, data.Email); err != nil {
		handleError(w, err)
		return
	}

	serverGenerateJSONResponse(w, errorResp, http.StatusOK)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := decodeData(w, r.Body, &userData); err != nil {
		return
	}

	errorResp := serverResponse{
		Error: "OK",
	}

	statusCode := http.StatusOK

	if err := user.AddUser(userData); err != nil {
		errorResp.Error = err.Error()
		if err == user.ErrEmailAlreadyTaken || err == user.ErrUsernameAlreadyTaken {
			statusCode = http.StatusConflict
		} else {
			statusCode = http.StatusBadRequest
		}
	}

	serverGenerateJSONResponse(w, errorResp, statusCode)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
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

func UserDataHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		handleError(w, err)
		return
	}

	claims, err := auth.ParseJWTToken(token.Value)
	if err != nil {
		handleError(w, err)
		return
	}

	userData, err := user.GetUserPublicInfo(claims.Email)
	response := serverResponse{
		PublicData: &userData,
	}

	if err != nil {
		handleError(w, err)
		return
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}
