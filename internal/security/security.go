package security

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

var cost int = 10

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

func ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomHash() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

