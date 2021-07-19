package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// SecureKey is a []byte used to encrypt/decrypt strings
type SecureKey []byte

func GenerateSecureKey(bytes int) *SecureKey {
	key := make(SecureKey, bytes)

	_, err := rand.Read(key)
	if err != nil {
		panic("failed to generate random key")
	}

	return &key
}

func (sk SecureKey) Base64String() string {
	return base64.URLEncoding.EncodeToString(sk)
}

func (sk *SecureKey) MarshalYAML() (interface{}, error) {
	return sk.Base64String(), nil
}

func (sk *SecureKey) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var key string
	if err := unmarshal(&key); err != nil {
		return err
	}

	decoded, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	*sk = make(SecureKey, len(decoded))
	copy(*sk, decoded)
	return nil
}

// Encrypt string to base64 crypto using AES
func Encrypt(text string, key []byte) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt from base64 to decrypted string
func Decrypt(cryptoText string, key []byte) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}
