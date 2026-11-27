package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

func Random(size int) string {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func Digest(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:] }

func Seal(key []byte, value, aad string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, []byte(value), []byte(aad)), nil
}

func Unseal(key, data []byte, aad string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted material")
	}
	plain, err := gcm.Open(nil, data[:gcm.NonceSize()], data[gcm.NonceSize():], []byte(aad))
	return string(plain), err
}

func UpstreamAAD(id int64, address string) string {
	return fmt.Sprintf("upstream:v2:%d:%s", id, address)
}
