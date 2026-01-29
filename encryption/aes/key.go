package aes

import (
	"crypto/rand"
	"crypto/sha256"
)

func DeriveKeyFromString(password string) (key []byte) {

	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

func GenerateKey() (key []byte, err error) {

	key = make([]byte, 32)
	if _, err = rand.Read(key); err != nil {
		return
	}
	return
}
