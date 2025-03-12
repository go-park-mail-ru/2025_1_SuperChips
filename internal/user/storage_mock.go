package user

// [АК] Временное решение, чтобы организовать тесты, позже отрефакторить и удалить.
func (storage MapUserStorage) SetUserID(ID uint64, email string) {
	user, found := storage.findUserByMail(email)
	if !found {
		return
	}

	user.Id = 0
	storage.users[email] = user
}
