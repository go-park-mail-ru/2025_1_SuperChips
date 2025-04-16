package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/csrf"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
)

type UserUsecaseInterface interface {
	AddUser(user domain.User) (uint64, error)
	LoginUser(email, password string) (uint64, error)
}
  
type AuthHandler struct {
	Config      configs.Config
	UserService UserUsecaseInterface
	JWTManager  auth.JWTManager
}

type CSRFResponse struct {
	CSRFToken string `json:"csrf_token"`
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

	id, err := app.UserService.LoginUser(data.Email, data.Password)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	if err := app.setCookieJWT(w, app.Config, data.Email, id); err != nil {
		handleAuthError(w, err)
		return
	}

	token, err := csrf.GenerateCSRF()
	if err != nil {
		handleAuthError(w, err)
		return
	}

	app.setCookieCSRF(w, app.Config, token)

	csrfData := CSRFResponse{
		CSRFToken: token,
	}

	response.Description = "OK"
	response.Data = csrfData
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
	type registerData struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var regData registerData

	if err := DecodeData(w, r.Body, &regData); err != nil {
		return
	}

	response := ServerResponse{
		Description: "Created",
	}

	statusCode := http.StatusCreated

	userData := domain.User{
		Username: regData.Username,
		Password: regData.Password,
		Email:    regData.Email,
	}

	id, err := app.UserService.AddUser(userData)
	if err != nil {
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

	if err := app.setCookieJWT(w, app.Config, userData.Email, id); err != nil {
		handleAuthError(w, err)
		return
	}

	token, err := csrf.GenerateCSRF()
	if err != nil {
		handleAuthError(w, err)
		return
	}

	app.setCookieCSRF(w, app.Config, token)

	data := CSRFResponse{
		CSRFToken: token,
	}

	response.Data = data

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
	setCookie(w, changedConfig, csrf.CSRFToken, "", true)

	response := ServerResponse{
		Description: "logged out",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}


func setCookie(w http.ResponseWriter, config configs.Config, name string, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		// Domain:   "yourflow.ru",
		Path:     "/",
		HttpOnly: httpOnly,
		Secure:   config.CookieSecure,
		SameSite: http.SameSiteLaxMode,
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

func (app AuthHandler) setCookieCSRF(w http.ResponseWriter, config configs.Config, token string) {
	setCookie(w, config, csrf.CSRFToken, token, true)
}

func CheckAuth(r *http.Request, manager auth.JWTManager) (*auth.Claims, error) {
	token, err := r.Cookie(auth.AuthToken)
	if err != nil {
		return nil, err
	}

	claims, err := manager.ParseJWTToken(token.Value)
	if err != nil {
		return nil, err
	}

	return claims, nil
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
	case errors.Is(err, auth.ErrorExpiredToken):
		errorResp.Description = http.StatusText(http.StatusUnauthorized)
		ServerGenerateJSONResponse(w, errorResp, http.StatusUnauthorized)
	default:
		ServerGenerateJSONResponse(w, errorResp, http.StatusInternalServerError)
	}
}
