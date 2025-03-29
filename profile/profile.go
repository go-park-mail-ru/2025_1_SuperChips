package profile

import "github.com/go-park-mail-ru/2025_1_SuperChips/domain"

type ProfileRepository interface {
	GetUserPublicInfoByEmail(email string) (domain.User, error)
	GetUserPublicInfoByUsername(username string) (domain.User, error)
	SaveUserAvatar(email string, avatar string) error
	UpdateUserData(user domain.User) error
}

type ProfileService struct {
	repo ProfileRepository
}

func NewProfileService(repo ProfileRepository) *ProfileService {
	return &ProfileService{
		repo: repo,
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

func (p *ProfileService) UpdateUserData(user domain.User) error {
	if err := user.ValidateUser(); err != nil {
		return err
	}

	err := p.repo.UpdateUserData(user)
	if err != nil {
		return err
	}

	return nil
}
