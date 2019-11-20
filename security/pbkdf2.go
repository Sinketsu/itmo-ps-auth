package security

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

func HashPassword(password string) string {
	saltBytes := make([]byte, 16)
	rand.Read(saltBytes)
	saltString := base64.StdEncoding.EncodeToString(saltBytes)
	hash := pbkdf2.Key([]byte(password), saltBytes, 8, 32, sha512.New384)
	hashEncoded := base64.StdEncoding.EncodeToString(hash)
	return saltString + "$" + hashEncoded
}

func VerifyPassword(password string, hash string) bool {
	splitted := strings.Split(hash, "$")
	saltBytes, _ := base64.StdEncoding.DecodeString(splitted[0])
	hashed := pbkdf2.Key([]byte(password), saltBytes, 8, 32, sha512.New384)
	hashedEncoded := base64.StdEncoding.EncodeToString(hashed)

	return hashedEncoded == splitted[1]
}
