package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)

func setCookie(w http.ResponseWriter, config configs.Config, name string, value string, httpOnly bool) {
    http.SetCookie(w, &http.Cookie{
        Name:     auth.AuthToken,
        Value:    value,
        Path:     "/",
        HttpOnly: httpOnly,
        Secure:   config.CookieSecure,
        SameSite: http.SameSiteLaxMode,
        Expires:  time.Now().Add(config.ExpirationTime),
    })
}

func setCookieJWT(w http.ResponseWriter, config configs.Config, email string, userID uint64) error {
    tokenString, err := auth.CreateJWT(config, userID, email)
    if err != nil {
        return err
    }

    setCookie(w, config, auth.AuthToken, tokenString, true)

    return nil
}


func (app AppHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	data := loginData{}
	if err := decodeData(w, r.Body, &data); err != nil {
		return
	}

	response := serverResponse{
		Description: "OK",
	}

	if err := app.UserStorage.LoginUser(data.Email, data.Password); err != nil {
		response.Description = "invalid credentials"
		serverGenerateJSONResponse(w, response, http.StatusForbidden)
		return
	}

	id := app.UserStorage.GetUserId(data.Email)

	if err := setCookieJWT(w, app.Config, data.Email, id); err != nil {
		handleError(w, err)
		return
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

func (app AppHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := decodeData(w, r.Body, &userData); err != nil {
		return
	}

	response := serverResponse{
		Description: "OK",
	}

	statusCode := http.StatusCreated

	if err := userData.ValidateUser(); err != nil {
		switch {
		case errors.Is(err, user.ErrValidation):
			statusCode = http.StatusBadRequest
			switch {
			case errors.Is(err, user.ErrInvalidBirthday):
				response.Description = "Invalid birthday"
			case errors.Is(err, user.ErrNoPassword):
				response.Description = "Password not given"
			case errors.Is(err, user.ErrInvalidEmail):
				response.Description = "Invalid email"
			case errors.Is(err, user.ErrInvalidUsername):
				response.Description = "Invalid username"
			case errors.Is(err, user.ErrPasswordTooLong):
				response.Description = "Password is too long"
			default:
				response.Description = "Bad request"
			}
		}
	}

	if err := app.UserStorage.AddUser(userData); err != nil {
		switch {
		case errors.Is(err, user.ErrConflict):
			statusCode = http.StatusConflict
			response.Description = "This email or username is already used"
		default:
			handleError(w, err)
			return
		}
	}

	serverGenerateJSONResponse(w, response, statusCode)
}

func (app AppHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	changedConfig := app.Config
	changedConfig.ExpirationTime = -time.Hour * 24 * 365

	setCookie(w, changedConfig, auth.AuthToken, "", true)

	response := serverResponse{
		Description: "logged out",
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

func (app AppHandler) UserDataHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		handleError(w, err)
		return
	}

	claims, err := auth.ParseJWTToken(token.Value, app.Config)
	if err != nil {
		if errors.Is(err, auth.ErrorExpiredToken) {
			handleHttpError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		handleError(w, err)
		return
	}

	userData, err := app.UserStorage.GetUserPublicInfo(claims.Email)
	if err != nil {
		handleError(w, err)
		return
	}
	response := serverResponse{
		Data: &userData,
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

