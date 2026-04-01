package utils

import (
	"crypto"
	"encoding/hex"
)

func HashPassword(password string) string {
	hash_function := crypto.SHA256.New()
	hash_function.Write([]byte(password))
	password_hashed := hash_function.Sum(nil)
	return hex.EncodeToString(password_hashed)
}
