package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/cookie"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
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

	w.Header().Set("content-type", "application/json")

	errorMap := make(map[string]string)
	errorMap["error"] = "OK"

	if err := user.LoginUser(data.Email, data.Password); err != nil {
		errorMap["error"] = err.Error()
		w.WriteHeader(http.StatusForbidden)
		if err := json.NewEncoder(w).Encode(errorMap); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if err := cookie.CookieSetJWT(w, data.Email); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(errorMap); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	errorMap := make(map[string]string)
	errorMap["error"] = "OK"

	if err := user.AddUser(userData); err != nil {
		errorMap["error"] = err.Error()
		w.WriteHeader(http.StatusBadRequest)

		if err == user.ErrEmailAlreadyTaken || err == user.ErrUsernameAlreadyTaken {
			w.WriteHeader(http.StatusConflict)
		}
		if err == user.ErrInternalError {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(errorMap); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	if err := json.NewEncoder(w).Encode(errorMap); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
    expiredCookie := &http.Cookie{
        Name:     "auth_token",
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
	token, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	claims, err := cookie.ParseJWTToken(token.Value)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	userData, err := user.GetUserPublicInfo(claims.Email)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("content-type", "application/json")

	if err := json.NewEncoder(w).Encode(userData); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
