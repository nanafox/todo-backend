package utils

import "golang.org/x/crypto/bcrypt"

// VerifyPasswordHash verifies that a password hash provided matches an existing hash
func VerifyPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))

	return err == nil
}
