package security

import "golang.org/x/crypto/bcrypt"

// Hash hashes data for security purpose. For example, not to store raw unprotected data on database.
func Hash(value string, hashCost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(value), hashCost)
	return string(bytes), err
}

// ValidateHash checks if hashed data is equal to raw data.
func ValidateHash(value, hashedValue string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedValue), []byte(value))
	return err == nil
}
