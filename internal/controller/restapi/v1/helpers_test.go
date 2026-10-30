package v1

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var errInternal = errors.New("internal error")

func bcryptHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash), err
}
