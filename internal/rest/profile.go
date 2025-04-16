package rest

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

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

type passwordChange struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UserUpdateRequest struct {
    Username   *string    `json:"username"`
    Email      *string    `json:"email"`
    Birthday   *time.Time `json:"birthday"`
    About      *string    `json:"about"`
    PublicName *string    `json:"public_name"`
}

type ProfileService interface {
	GetUserPublicInfoByEmail(email string) (domain.User, error)
	GetUserPublicInfoByUsername(username string) (domain.User, error)
	SaveUserAvatar(email string, avatar string) error
	UpdateUserData(user domain.User, oldEmail string) error
	ChangeUserPassword(email, oldPassword, newPassword string) (int, error)
}

type ProfileHandler struct {
	ProfileService ProfileService
	JwtManager     auth.JWTManager
	AvatarFolder   string        // где будут хранится аватары относительно staticFolder
	StaticFolder   string        // где будут хранится статические файлы
	BaseUrl        string        // url для получения аватара
	ExpirationTime time.Duration // время жизни куки
	CookieSecure   bool          // флаг, что куки должны быть только по https
}

func (h *ProfileHandler) CurrentUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok {
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
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
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
		HttpErrorToJson(w, "image is too large", http.StatusRequestEntityTooLarge)
		return
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	detected := http.DetectContentType(buffer)
	contentType := handler.Header.Get("Content-Type")

	if !strings.HasPrefix(detected, strings.Split(contentType, ";")[0]) {
		HttpErrorToJson(w, "image extension and type are mismatched", http.StatusBadRequest)
		return
	}

	if _, ok := allowedTypes[detected]; !ok {
		HttpErrorToJson(w, "this extension is not supported", http.StatusBadRequest)
		return
	}

	if _, err := file.Seek(0, 0); err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	filename := filepath.Base(handler.Filename)
    ext := filepath.Ext(filename)
    if ext == "" {
        HttpErrorToJson(w, "invalid file extension", http.StatusBadRequest)
        return
    }

	filename, url, err := image.UploadImage(handler.Filename, h.StaticFolder, h.AvatarFolder, h.BaseUrl, file)
	if err != nil {
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = h.ProfileService.SaveUserAvatar(claims.Email, filename)
	if err != nil {
		handleProfileError(w, err)
		return
	}

	type imageURL struct {
		MediaURL string `json:"media_url"`
	}

	imgURL := imageURL{
		MediaURL: url,
	}

	response := ServerResponse{
		Description: "Created",
		Data: imgURL,
	}

	ServerGenerateJSONResponse(w, response, http.StatusCreated)
}

func (h *ProfileHandler) ChangeUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	var passwordStruct passwordChange

	if err := DecodeData(w, r.Body, &passwordStruct); err != nil {
		return
	}

	id, err := h.ProfileService.ChangeUserPassword(claims.Email, passwordStruct.OldPassword, passwordStruct.NewPassword)
	if err != nil {
		handleProfileError(w, err)
		return
	}

	conf := configs.Config{
		ExpirationTime: h.ExpirationTime,
		CookieSecure:   h.CookieSecure,
	}

	if err := updateAuthToken(w, h.JwtManager, conf, claims.Email, id); err != nil {
		handleProfileError(w, err)
		return
	}

	resp := ServerResponse{
		Description: "OK",
	}

	ServerGenerateJSONResponse(w, resp, http.StatusOK)
}


func (h *ProfileHandler) PatchUserProfileHandler(w http.ResponseWriter, r *http.Request) {
    claims, ok := r.Context().Value(auth.ClaimsContextKey).(*auth.Claims)
	if !ok || claims == nil {
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

    var updateReq UserUpdateRequest
    if err := DecodeData(w, r.Body, &updateReq); err != nil {
        return
    }

	existingUser, err := h.ProfileService.GetUserPublicInfoByEmail(claims.Email)
    if err != nil {
        handleProfileError(w, err)
        return
    }

    if updateReq.Username != nil {
        existingUser.Username = *updateReq.Username
    }
    if updateReq.Email != nil {
        existingUser.Email = *updateReq.Email
    }
    if updateReq.Birthday != nil {
        existingUser.Birthday = *updateReq.Birthday
    }
    if updateReq.About != nil {
        existingUser.About = *updateReq.About
    }
    if updateReq.PublicName != nil {
        existingUser.PublicName = *updateReq.PublicName
    }

    if err := existingUser.ValidateUserNoPassword(); err != nil {
        HttpErrorToJson(w, "validation failed", http.StatusBadRequest)
        return
    }

    if err := h.ProfileService.UpdateUserData(existingUser, claims.Email); err != nil {
        handleProfileError(w, err)
        return
    }

    if updateReq.Email != nil && *updateReq.Email != claims.Email {
        conf := configs.Config{
            ExpirationTime: h.ExpirationTime,
            CookieSecure:   h.CookieSecure,
        }
        if err := updateAuthToken(w, h.JwtManager, conf, existingUser.Email, int(existingUser.Id)); err != nil {
            handleProfileError(w, err)
            return
        }
    }

    ServerGenerateJSONResponse(w, ServerResponse{Description: "OK"}, http.StatusOK)
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
	case errors.Is(err, domain.ErrNoPassword):
		HttpErrorToJson(w, "cannot use empty password", http.StatusBadRequest)
	case errors.Is(err, domain.ErrPasswordTooLong):
		HttpErrorToJson(w, "password too long", http.StatusBadRequest)
	case errors.Is(err, http.ErrNoCookie):
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	case errors.Is(err, auth.ErrorExpiredToken):
		HttpErrorToJson(w, "session expired", http.StatusUnauthorized)
	case errors.Is(err, domain.ErrInvalidCredentials):
		HttpErrorToJson(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	case errors.Is(err, io.EOF):
		HttpErrorToJson(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	default:
		HttpErrorToJson(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

}

func updateAuthToken(w http.ResponseWriter, mngr auth.JWTManager, config configs.Config, email string, id int) error {
	token, err := mngr.CreateJWT(email, id)
	if err != nil {
		return err
	}
	setCookie(w, config, auth.AuthToken, token, true)

	return nil
}
