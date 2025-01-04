package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword hashes a plain text password
func EncryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePasswords compares a hashed password with a plain text password
func ComparePasswords(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.New("password does not match")
	}
	return nil
}
