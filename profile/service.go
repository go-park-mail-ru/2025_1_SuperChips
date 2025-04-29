package profile

import (
	"path/filepath"

	"github.com/go-park-mail-ru/2025_1_SuperChips/domain"
	"github.com/go-park-mail-ru/2025_1_SuperChips/internal/security"
)

type ProfileRepository interface {
	GetUserPublicInfoByEmail(email string) (domain.User, error)
	GetUserPublicInfoByUsername(username string) (domain.User, error)
	SaveUserAvatar(email string, avatar string) error
	UpdateUserData(user domain.User, oldEmail string) error
	GetHashedPassword(email string) (string, error)
	SetNewPassword(email string, newPassword string) (int, error)
}

type ProfileService struct {
	repo      ProfileRepository
	baseURL   string
	staticDir string
	avatarDir string
}

func NewProfileService(repo ProfileRepository, baseURL, staticDir, avatarDir string) *ProfileService {
	return &ProfileService{
		repo: repo,
		baseURL: baseURL,
		staticDir: staticDir,
		avatarDir: avatarDir,
	}
}

func (p *ProfileService) GetUserPublicInfoByEmail(email string) (domain.User, error) {
	if err := domain.ValidateEmail(email); err != nil {
		return domain.User{}, err
	}

	user, err := p.repo.GetUserPublicInfoByEmail(email)
	if err != nil {
		return domain.User{}, err
	}

	if (!user.IsExternalAvatar) {
		user.Avatar = p.generateAvatarURL(user.Avatar)
	}

	return user, nil
}

func (p *ProfileService) GetUserPublicInfoByUsername(username string) (domain.User, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return domain.User{}, err
	}

	user, err := p.repo.GetUserPublicInfoByUsername(username)
	if err != nil {
		return domain.User{}, err
	}

	if (!user.IsExternalAvatar) {
		user.Avatar = p.generateAvatarURL(user.Avatar)
	}

	return user, nil
}

func (p *ProfileService) SaveUserAvatar(email string, avatar string) error {
	if err := domain.ValidateEmail(email); err != nil {
		return err
	}

	err := p.repo.SaveUserAvatar(email, avatar)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProfileService) UpdateUserData(user domain.User, oldEmail string) error {
	if err := user.ValidateUserNoPassword(); err != nil {
		return err
	}

	err := p.repo.UpdateUserData(user, oldEmail)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProfileService) ChangeUserPassword(email, oldPassword, newPassword string) (int, error) {
	if err := domain.ValidatePassword(newPassword); err != nil {
		return 0, err
	}

	hash, err := p.repo.GetHashedPassword(email)
	if err != nil {
		return 0, err
	}

	if !security.ComparePassword(oldPassword, hash) {
		return 0, domain.ErrInvalidCredentials
	}

	newHash, err := security.HashPassword(newPassword)
	if err != nil {
		return 0, err
	}

	id, err := p.repo.SetNewPassword(email, newHash)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *ProfileService) generateAvatarURL(filename string) string {
	if filename == "" {
		return ""
	}

	return p.baseURL + filepath.Join(p.staticDir, p.avatarDir, filename)
}
