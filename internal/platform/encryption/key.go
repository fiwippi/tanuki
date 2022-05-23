// Package encryption provides methods to encrypt and decrypt
// strings using a securely generated secret key
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Key represents a secret key used to encrypt/decrypt strings
type Key []byte

// NewKey generates a new secret key with the given amount of bytes
func NewKey(bytes int) *Key {
	key := make(Key, bytes)

	_, err := rand.Read(key)
	if err != nil {
		panic("failed to generate random key")
	}

	return &key
}

func (k Key) Base64() string {
	return base64.URLEncoding.EncodeToString(k)
}

// Encrypt encrypts a piece of text using the secret key
func (k Key) Encrypt(text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(k)
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

	// Convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}

// Decrypt decrypts a piece of text using the secret key
func (k Key) Decrypt(cryptoText string) string {
	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(k)
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

func (k *Key) MarshalYAML() (interface{}, error) {
	return k.Base64(), nil
}

func (k *Key) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var key string
	if err := unmarshal(&key); err != nil {
		return err
	}

	decoded, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	*k = make(Key, len(decoded))
	copy(*k, decoded)
	return nil
}
