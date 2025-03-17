package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/errs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/user"
)


type loginData struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

// LoginHandler godoc
// @Summary Log in user
// @Description Tries to log the user in
// @Accept json
// @Produce json
// @Param email body string true "user email" example("user@mail.ru")
// @Param password body string true "user password" example("abcdefgh1234")
// @Success 200 string Description "OK"
// @Failure 400 string Description "Bad Request"
// @Failure 403 string Description "invalid credentials"
// @Failure 500 string Description "Internal server error"
// @Router /api/v1/auth/login [post]
func (app AppHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var data loginData
	if err := decodeData(w, r.Body, &data); err != nil {
		return
	}

	var response serverResponse

	if err := user.ValidateEmailAndPassword(data.Email, data.Password); err != nil {
		handleAuthError(w, err)
		return
	}

	if err := app.UserStorage.LoginUser(data.Email, data.Password); err != nil {
		handleAuthError(w, err)
		return
	}

	id := app.UserStorage.GetUserId(data.Email)

	if err := app.setCookieJWT(w, app.Config, data.Email, id); err != nil {
		handleAuthError(w, err)
		return
	}

	response.Description = "OK"
	serverGenerateJSONResponse(w, response, http.StatusOK)
}

// RegistrationHandler godoc
// @Summary Register user
// @Description Tries to register the user
// @Accept json
// @Produce json
// @Param email body string true "user email" example("admin@mail.ru")
// @Param username body string true "user username" example("mailrudabest")
// @Param birthday body string true "user date of birth RFC" example("1990-12-31T23:59:60Z")
// @Param password body string true "user password" example("unbreakable_password")
// @Success 201 string serverResponse.Description "Created"
// @Failure 400 string serverResponse.Description "Bad Request"
// @Failure 409 string serverResponse.Description "Conflict"
// @Failure 500 string serverResponse.Description "Internal server error"
// @Router /api/v1/auth/register [post]
func (app AppHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	userData := user.User{}
	if err := decodeData(w, r.Body, &userData); err != nil {
		return
	}

	response := serverResponse{
		Description: "Created",
	}

	statusCode := http.StatusCreated

	if err := userData.ValidateUser(); err != nil {
		switch {
		case errors.Is(err, errs.ErrValidation):
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
		default: 
			handleAuthError(w, err)
			return
		}

		serverGenerateJSONResponse(w, response, statusCode)
		return
	}

	if err := app.UserStorage.AddUser(userData); err != nil {
		switch {
		case errors.Is(err, errs.ErrConflict):
			statusCode = http.StatusConflict
			response.Description = "This email or username is already used"
		default:
			handleAuthError(w, err)
			return
		}
	}

	serverGenerateJSONResponse(w, response, statusCode)
}

// LogoutHandler godoc
// @Summary Logout user
// @Description Logouts user
// @Produce json
// @Success 200 string serverResponse.Description "logged out"
// @Router /api/v1/auth/logout [post]
func (app AppHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	changedConfig := app.Config
	changedConfig.ExpirationTime = -time.Hour * 24 * 365

	setCookie(w, changedConfig, auth.AuthToken, "", true)

	response := serverResponse{
		Description: "logged out",
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

// UserDataHandler godoc
// @Summary Get user data
// @Description Tries to get current user's data
// @Produce json
// @Success 200 body serverResponse.Data
// @Failure 400 string serverResponse.Description "Bad Request"
// @Failure 401 string serverResponse.Description "Unauthorized"
// @Failure 500 string serverResponse.Description "Internal server error"
// @Router /api/v1/auth/user [get]
func (app AppHandler) UserDataHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	claims, err := app.JWTManager.ParseJWTToken(token.Value)
	if err != nil {
		if errors.Is(err, auth.ErrorExpiredToken) {
			httpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		handleAuthError(w, err)
		return
	}

	userData, err := app.UserStorage.GetUserPublicInfo(claims.Email)
	if err != nil {
		handleAuthError(w, err)
		return
	}
	response := serverResponse{
		Data: &userData,
	}

	serverGenerateJSONResponse(w, response, http.StatusOK)
}

func setCookie(w http.ResponseWriter, config configs.Config, name string, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Now().Add(config.ExpirationTime),
	})
}

func (app AppHandler) setCookieJWT(w http.ResponseWriter, config configs.Config, email string, userID uint64) error {
	tokenString, err := app.JWTManager.CreateJWT(email, int(userID))
	if err != nil {
		return err
	}

	setCookie(w, config, auth.AuthToken, tokenString, true)

	return nil
}

