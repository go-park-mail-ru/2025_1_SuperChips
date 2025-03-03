package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/cookie"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type errorAnswer struct {
	Error string `json:"error"`
}

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

func serverGenerateAnswer[T any](w http.ResponseWriter, body T) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("server is up"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := loginData{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

	if err := cookie.CookieAddJWT(w, data.Email); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	serverGenerateAnswer(w, errorResp)
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	errorResp := errorAnswer{
		Error: "OK",
	}

	if err := user.AddUser(userData); err != nil {
		errorResp.Error = err.Error()

		switch err {
		case user.ErrEmailAlreadyTaken, user.ErrUsernameAlreadyTaken:
			w.WriteHeader(http.StatusConflict)
		case user.ErrInternalError:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
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
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	claims, err := auth.ParseJWTToken(token.Value)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userData, err := user.GetUserPublicInfo(claims.Email)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	serverGenerateAnswer(w, userData)
}

