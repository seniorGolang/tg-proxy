package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"

	"github.com/seniorGolang/tg-proxy/errs"
	"github.com/seniorGolang/tg-proxy/helpers"
)

const keySize = 32
const nonceSize = 12

type Encryptor struct {
	key   []byte
	block cipher.Block
}

// NewEncryptor создает новый AES-GCM Encryptor
func NewEncryptor(key []byte) (enc *Encryptor, err error) {

	if len(key) != keySize {
		err = errs.ErrInvalidKeySize
		return
	}

	var block cipher.Block
	if block, err = aes.NewCipher(key); err != nil {
		return
	}

	enc = &Encryptor{
		key:   key,
		block: block,
	}
	return
}

func (e *Encryptor) Encrypt(plaintext []byte) (ciphertext []byte, err error) {

	nonce := make([]byte, nonceSize)
	if _, err = rand.Read(nonce); err != nil {
		return
	}

	var aesgcm cipher.AEAD
	if aesgcm, err = cipher.NewGCM(e.block); err != nil {
		return
	}

	encrypted := aesgcm.Seal(nil, nonce, plaintext, nil)
	ciphertext = make([]byte, 0, len(nonce)+len(encrypted))
	ciphertext = append(ciphertext, nonce...)
	ciphertext = append(ciphertext, encrypted...)

	return
}

func (e *Encryptor) Decrypt(ciphertext []byte) (plaintext []byte, err error) {

	if len(ciphertext) < nonceSize {
		plaintext = nil
		err = errs.ErrInvalidCiphertext
		return
	}

	nonce := ciphertext[:nonceSize]
	encryptedData := ciphertext[nonceSize:]

	var aesgcm cipher.AEAD
	if aesgcm, err = cipher.NewGCM(e.block); err != nil {
		return
	}

	if plaintext, err = aesgcm.Open(nil, nonce, encryptedData, nil); err != nil {
		plaintext = nil
		err = fmt.Errorf("%w: %v", errs.ErrDecryptionFailed, err)
		return
	}

	return
}

func (e *Encryptor) EncryptString(plaintext string) (encrypted string, err error) {

	var ciphertext []byte
	if ciphertext, err = e.Encrypt([]byte(plaintext)); err != nil {
		slog.Debug("Encryption failed",
			slog.String(helpers.LogKeyEncryptionType, "aes-gcm"),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	encrypted = base64.StdEncoding.EncodeToString(ciphertext)

	return
}

func (e *Encryptor) DecryptString(ciphertext string) (plaintext string, err error) {

	var data []byte
	if data, err = base64.StdEncoding.DecodeString(ciphertext); err != nil {
		slog.Debug("Failed to decode base64",
			slog.String(helpers.LogKeyEncryptionType, "aes-gcm"),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}

	var p []byte
	if p, err = e.Decrypt(data); err != nil {
		slog.Debug("Decryption failed",
			slog.String(helpers.LogKeyEncryptionType, "aes-gcm"),
			slog.Any(helpers.LogKeyError, err),
		)
		return
	}
	plaintext = string(p)

	return
}
