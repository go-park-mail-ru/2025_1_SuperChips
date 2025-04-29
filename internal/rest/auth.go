package rest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/csrf"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	gen "github.com/go-park-mail-ru/2025_1_SuperChips/protos/gen/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	Config          configs.Config
	UserService     gen.AuthClient
	JWTManager      auth.JWTManager
	ContextDuration time.Duration
}

type ExternalData struct {
	AccessToken string `json:"access_token,omitempty"`
	Username    string `json:"username,omitempty"`
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

	ctx, cancel := context.WithTimeout(context.Background(), app.ContextDuration)
	defer cancel()

	var response ServerResponse

	grpcResp, err := app.UserService.LoginUser(ctx, &gen.LoginUserRequest{
		Email: data.Email,
		Password: data.Password,
	})
	if err != nil {
		handleGRPCAuthError(w, err)
		return
	}

	if err := app.setCookieJWT(w, app.Config, data.Email, uint64(grpcResp.ID)); err != nil {
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

	ctx, cancel := context.WithTimeout(context.Background(), app.ContextDuration)
	defer cancel()

	grpcResp, err := app.UserService.AddUser(ctx, &gen.AddUserRequest{
		Email: userData.Email,
		Password: userData.Password,
		Username: userData.Username,
	})
	if err != nil {
		handleGRPCAuthError(w, err)
		return
	}

	if err := app.setCookieJWT(w, app.Config, userData.Email, uint64(grpcResp.ID)); err != nil {
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

func (app AuthHandler) ExternalLogin(w http.ResponseWriter, r *http.Request) {
	var data ExternalData
	if err := DecodeData(w, r.Body, &data); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.ContextDuration)
	defer cancel()

	VKid, VKemail, err := vkGetData(data.AccessToken, app.Config.VKClientID)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	grpcResp, err := app.UserService.LoginExternalUser(ctx, &gen.LoginExternalUserRequest{
		Email: VKemail,
		ExternalID: VKid,
	})
	if err != nil {
		handleGRPCAuthError(w, err)
		return
	}

	if err := app.setCookieJWT(w, app.Config, grpcResp.Email, uint64(grpcResp.ID)); err != nil {
		handleAuthError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	token, err := csrf.GenerateCSRF()
	if err != nil {
		handleAuthError(w, err)
		return
	}

	app.setCookieCSRF(w, app.Config, token)

	csrf := CSRFResponse{
		CSRFToken: token,
	}

	resp.Data = csrf

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func (app AuthHandler) ExternalRegister(w http.ResponseWriter, r *http.Request) {
	var data ExternalData
	if err := DecodeData(w, r.Body, &data); err != nil {
		return
	}

	VKid, VKemail, err := vkGetData(data.AccessToken, app.Config.VKClientID)
	if err != nil {
		handleAuthError(w, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), app.ContextDuration)
	defer cancel()

	grpcResp, err := app.UserService.AddExternalUser(ctx, &gen.AddExternalUserRequest{
		Email: VKemail,
		Username: data.Username,
		ExternalID: VKid,
	})
	if err != nil {
		handleGRPCAuthError(w, err)
		return
	}

	if err := app.setCookieJWT(w, app.Config, VKemail, uint64(grpcResp.ID)); err != nil {
		handleAuthError(w, err)
		return
	}

	token, err := csrf.GenerateCSRF()
	if err != nil {
		handleAuthError(w, err)
		return
	}

	app.setCookieCSRF(w, app.Config, token)

	csrf := CSRFResponse{
		CSRFToken: token,
	}

	resp := ServerResponse{
		Description: "OK",
		Data: csrf,
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}

func vkGetData(accessToken string, clientID string) (string, string, error) {
	type VKUser struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}

	type VKUserTop struct {
		User VKUser `json:"user"`
	}

	postURL := "https://id.vk.com/oauth2/user_info"

	formData := url.Values{}
	formData.Set("client_id", clientID)
	formData.Set("access_token", accessToken)

	req, err := http.NewRequest("POST", postURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var data VKUserTop
	if err := DecodeData(nil, resp.Body, &data); err != nil {
		return "", "", err
	}

	return data.User.UserID, data.User.Email, nil
}

func setCookie(w http.ResponseWriter, config configs.Config, name string, value string, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		// domain:   "yourflow.ru",
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
	case errors.Is(err, domain.ErrConflict):
		errorResp.Description = "account with that username or email already exists"
		ServerGenerateJSONResponse(w, errorResp, http.StatusConflict)
	default:
		ServerGenerateJSONResponse(w, errorResp, http.StatusInternalServerError)
	}
}

func handleGRPCAuthError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.Internal:
			HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		case codes.AlreadyExists:
			HttpErrorToJson(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		case codes.PermissionDenied:
			HttpErrorToJson(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		case codes.Unauthenticated:
			HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		case codes.NotFound:
			HttpErrorToJson(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		case codes.InvalidArgument:
			HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
