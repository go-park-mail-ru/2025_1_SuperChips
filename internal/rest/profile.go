package rest

import (
	"errors"
	"mime/multipart"
	"net/http"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	auth "github.com/go-park-mail-ru/2025_1_SuperChips/internal/rest/auth"
	"github.com/go-park-mail-ru/2025_1_SuperChips/utils/image"
)

const maxAvatarSize = (1 << 20) * 3 // 3 мб

var allowedTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/bmp":  true,
	"image/tiff": true,
}

type ProfileService interface {
	GetUserPublicInfoByEmail(email string) (domain.User, error)
	GetUserPublicInfoByUsername(username string) (domain.User, error)
	SaveUserAvatar(email string, avatar string) error
	UpdateUserData(user domain.User, oldEmail string) error
}

type ProfileHandler struct {
	ProfileService ProfileService
	JwtManager     auth.JWTManager
	Config         configs.Config
	AvatarFolder   string // где будут хранится аватары относительно staticFolder
	StaticFolder   string // где будут хранится статические файлы
	BaseUrl        string // url для получения аватара
}

func (h *ProfileHandler) CurrentUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		h.updateUserProfileHandler(w, r)
		return
	}

	cookie, err := r.Cookie(auth.AuthToken)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token := cookie.Value

	claims, err := h.JwtManager.ParseJWTToken(token)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user, err := h.ProfileService.GetUserPublicInfoByEmail(claims.Email)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := ServerResponse{
		Data: user,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func (h *ProfileHandler) PublicProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		HttpErrorToJson(w, "username is empty", http.StatusBadRequest)
		return
	}

	user, err := h.ProfileService.GetUserPublicInfoByUsername(username)
	if err != nil {
		handleProfileError(w, err)
		return
	}

	response := ServerResponse{
		Data: user,
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func (h *ProfileHandler) UserAvatarHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := CheckAuth(r, h.JwtManager)
	if errors.Is(err, auth.ErrorExpiredToken) {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else if err != nil {
		handleProfileError(w, err)
		return
	}

	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		handleProfileError(w, err)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		handleProfileError(w, err)
		return
	}
	defer file.Close()

	if handler.Size > maxAvatarSize {
		handleProfileError(w, multipart.ErrMessageTooLarge)
		return
	}

	if !allowedTypes[handler.Header.Get("Content-Type")] {
		HttpErrorToJson(w, "unsupported file format", http.StatusUnsupportedMediaType)
		return
	}

	url, err := image.UploadImage(handler.Filename, h.StaticFolder, h.AvatarFolder, h.BaseUrl, file)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.ProfileService.SaveUserAvatar(claims.Email, url)
	if err != nil {
		handleProfileError(w, err)
		return
	}

	response := ServerResponse{
		Description: "Created",
	}

	ServerGenerateJSONResponse(w, response, http.StatusCreated)
}

func (h *ProfileHandler) updateUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := CheckAuth(r, h.JwtManager)
	if errors.Is(err, auth.ErrorExpiredToken) {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	} else if err != nil {
		handleProfileError(w, err)
		return
	}

	bufUser, err := h.ProfileService.GetUserPublicInfoByEmail(claims.Email)
	if err != nil {
		handleProfileError(w, err)
		return
	}

	if err := DecodeData(w, r.Body, &bufUser); err != nil {
		handleProfileError(w, err)
		return
	}

	if err := bufUser.ValidateUserNoPassword(); err != nil {
		HttpErrorToJson(w, "validation failed", http.StatusBadRequest)
		return
	}

	user := domain.User{
		Username:   bufUser.Username,
		Email:      bufUser.Email,
		Birthday:   bufUser.Birthday,
		About:      bufUser.About,
		PublicName: bufUser.PublicName,
	}

	if err := h.ProfileService.UpdateUserData(user, claims.Email); err != nil {
		handleProfileError(w, err)
		return
	}

	response := ServerResponse{
		Description: "OK",
	}

	// в будущем!! 
	// обновить версию токена в бд,
	// тем самым обнулив все предыдущие токены
	token, err := h.JwtManager.CreateJWT(user.Email, int(bufUser.Id))
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	setCookie(w, h.Config, auth.AuthToken, token, true)

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func handleProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, multipart.ErrMessageTooLarge):
		HttpErrorToJson(w, "image is too large", http.StatusRequestEntityTooLarge)
	case errors.Is(err, domain.ErrUserNotFound):
		HttpErrorToJson(w, "user not found", http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidEmail):
		HttpErrorToJson(w, "invalid email", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidUsername):
		HttpErrorToJson(w, "invalid username", http.StatusBadRequest)
	case errors.Is(err, domain.ErrInvalidBirthday):
		HttpErrorToJson(w, "invalid birthday", http.StatusBadRequest)
	case errors.Is(err, http.ErrNoCookie):
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}
