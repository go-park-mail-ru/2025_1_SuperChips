package rest

import (
	"errors"
	"mime/multipart"
	"net/http"

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
	GetUserPublicInfo(email string) (domain.User, error)
	SaveUserAvatar(email string, avatar string) error
	UpdateUserData(user domain.User) error
}

type ProfileHandler struct {
	profileService ProfileService
	jwtManager     auth.JWTManager
	avatarFolder   string // где будут хранится аватары относительно staticFolder
	staticFolder   string
}

func NewProfileHandler(profileService ProfileService) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
	}
}

func (h *ProfileHandler) CurrentUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.AuthToken)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	token := cookie.Value

	claims, err := h.jwtManager.ParseJWTToken(token)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	user, err := h.profileService.GetUserPublicInfo(claims.Email)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ServerGenerateJSONResponse(w, user, http.StatusOK)
}

func (h *ProfileHandler) UserAvatarHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := image.UploadImage(handler.Filename, h.staticFolder, h.avatarFolder, file); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, response, http.StatusOK)
}

func handleProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, multipart.ErrMessageTooLarge):
		HttpErrorToJson(w, "image is too large", http.StatusRequestEntityTooLarge)
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}
