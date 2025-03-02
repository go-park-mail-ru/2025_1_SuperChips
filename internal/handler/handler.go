package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

var JWT_SECRET []byte = configs.Config.JWTSecret

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

	// здесь будет работа с куками
	// ...
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

	// работа с куками здесь
	// или можно не авторизовывать пользователя после регистрации,
	// а просить войти через /api/v1/auth/login
	// ...

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(errorMap); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// TODO
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// TODO
func UserDataHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}