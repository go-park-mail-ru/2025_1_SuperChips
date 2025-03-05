package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type errorAnswer struct {
	Error string `json:"error"`
}

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
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


func serverGenerateAnswer(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(body); err != nil {
		handleError(w, err)
	}
}

// здесь кмк без дженерика никак, так как *interface{} у меня не сработал
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

	errorResp := errorAnswer{
		Error: "OK",
	}

	if err := user.LoginUser(data.Email, data.Password); err != nil {
		errorResp.Error = err.Error()
		w.WriteHeader(http.StatusForbidden)
		serverGenerateAnswer(w, errorResp)
		return
	}

	if err := auth.CookieAddJWT(w, data.Email); err != nil {
		handleError(w, err)
		return
	}

	serverGenerateAnswer(w, errorResp)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := decodeData(w, r.Body, &userData); err != nil {
		return
	}

	errorResp := errorAnswer{
		Error: "OK",
	}

	if err := user.AddUser(userData); err != nil {
		errorResp.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	}
	
	serverGenerateAnswer(w, errorResp)
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
	if err != nil {
		handleError(w, err)
		return
	}

	serverGenerateAnswer(w, userData)
}

