package core

type encryptor interface {
	EncryptString(plaintext string) (ciphertext string, err error)
	DecryptString(ciphertext string) (plaintext string, err error)
}
