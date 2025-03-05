package user

type userRepository interface {
	containsUsername(username string) bool
	containsEmail(email string) bool
	findUserByMail(email string) (User, bool)
	addUserToBase(user User)
	initialize()
}

type userStorage struct {
	repo userRepository
}

type mapUserStorage struct {
	users map[string]User
}

func initUserStorage(repo userRepository) userStorage {
	repo.initialize()
	return userStorage{
		repo: repo,
	}
}

func (u *mapUserStorage) initialize() {
	u.users = make(map[string]User, 0)
}

func (u mapUserStorage) containsUsername(username string) bool {
	for _, v := range u.users {
		if v.Username == username {
			return true
		}
	}

	return false
}

func (u mapUserStorage) containsEmail(email string) bool {
	for _, v := range u.users {
		if v.Email == email {
			return true
		}
	}

	return false
}

func (u mapUserStorage) findUserByMail(email string) (User, bool) {
	for _, v := range u.users {
		if v.Email == email {
			return v, true
		}
	}

	return User{}, false
}

func (u *mapUserStorage) addUserToBase(user User) {
	u.users[user.Email] = user
}

