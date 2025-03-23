package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/user"
)

type AuthHandler struct {
	Config      configs.Config
	UserService user.UserService
	JWTManager  auth.JWTManager
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
func (app AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var data domain.LoginData
	if err := DecodeData(w, r.Body, &data); err != nil {
		return
	}

	var response ServerResponse

	if err := app.UserService.LoginUser(data.Email, data.Password); err != nil {
		handleAuthError(w, err)
		return
	}

	id := app.UserService.GetUserId(data.Email)

	if err := app.setCookieJWT(w, app.Config, data.Email, id); err != nil {
		handleAuthError(w, err)
		return
	}

	response.Description = "OK"
	ServerGenerateJSONResponse(w, response, http.StatusOK)
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
func (app AuthHandler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	var userData domain.User
	if err := DecodeData(w, r.Body, &userData); err != nil {
		return
	}

	response := ServerResponse{
		Description: "Created",
	}

	statusCode := http.StatusCreated

	if err := app.UserService.AddUser(userData); err != nil {
		switch {
		case errors.Is(err, domain.ErrValidation):
			statusCode = http.StatusBadRequest
			switch {
			case errors.Is(err, domain.ErrInvalidBirthday):
				response.Description = "Invalid birthday"
			case errors.Is(err, domain.ErrNoPassword):
				response.Description = "Password not given"
			case errors.Is(err, domain.ErrInvalidEmail):
				response.Description = "Invalid email"
			case errors.Is(err, domain.ErrInvalidUsername):
				response.Description = "Invalid username"
			case errors.Is(err, domain.ErrPasswordTooLong):
				response.Description = "Password is too long"
			default:
				response.Description = "Bad request"
			}
		case errors.Is(err, domain.ErrConflict):
			statusCode = http.StatusConflict
			response.Description = "This email or username is already used"
			ServerGenerateJSONResponse(w, response, statusCode)
			return
		default:
			handleAuthError(w, err)
			return
		}
	}

	id := app.UserService.GetUserId(userData.Email)

	if err := app.setCookieJWT(w, app.Config, userData.Email, id); err != nil {
		handleAuthError(w, err)
		return
	}

	ServerGenerateJSONResponse(w, response, statusCode)
}

// LogoutHandler godoc
// @Summary Logout user
// @Description Logouts user
// @Produce json
// @Success 200 string serverResponse.Description "logged out"
// @Router /api/v1/auth/logout [post]
func (app AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	changedConfig := app.Config
	changedConfig.ExpirationTime = -time.Hour * 24 * 365

	setCookie(w, changedConfig, auth.AuthToken, "", true)

	response := ServerResponse{
		Description: "logged out",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
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
func (app AuthHandler) UserDataHandler(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	claims, err := app.JWTManager.ParseJWTToken(token.Value)
	if err != nil {
		if errors.Is(err, auth.ErrorExpiredToken) {
			HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		handleAuthError(w, err)
		return
	}

	userData, err := app.UserService.GetUserPublicInfo(claims.Email)
	if err != nil {
		handleAuthError(w, err)
		return
	}
	response := ServerResponse{
		Data: &userData,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func setCookie(w http.ResponseWriter, config configs.Config, name string, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Domain:   "yourflow.ru",
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteNoneMode, // dont forget to change back to LAX when going to prod!!!!!!!!!1111
		Expires:  time.Now().Add(config.ExpirationTime),
	})
}

func (app AuthHandler) setCookieJWT(w http.ResponseWriter, config configs.Config, email string, userID uint64) error {
	tokenString, err := app.JWTManager.CreateJWT(email, int(userID))
	if err != nil {
		return err
	}

	setCookie(w, config, auth.AuthToken, tokenString, true)

	return nil
}

func handleAuthError(w http.ResponseWriter, err error) {
	var authErr domain.StatusError

	errorResp := ServerResponse{
		Description: http.StatusText(http.StatusInternalServerError),
	}

	switch {
	case errors.As(err, &authErr):
		errorResp.Description = authErr.Error()
		ServerGenerateJSONResponse(w, errorResp, authErr.StatusCode())
	case errors.Is(err, http.ErrNoCookie):
		errorResp.Description = http.StatusText(http.StatusForbidden)
		ServerGenerateJSONResponse(w, errorResp, http.StatusForbidden)
	case errors.Is(err, ErrBadRequest):
		errorResp.Description = http.StatusText(http.StatusBadRequest)
		ServerGenerateJSONResponse(w, errorResp, http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrValidation):
		errorResp.Description = "invalid credentials"
		ServerGenerateJSONResponse(w, errorResp, http.StatusUnauthorized)
	default:
		ServerGenerateJSONResponse(w, errorResp, http.StatusInternalServerError)
	}
}
