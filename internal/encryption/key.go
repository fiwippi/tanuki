// Package encryption provides methods to encrypt and decrypt
// strings using a securely generated secret key
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

// Key represents a secret key used to encrypt/decrypt strings
type Key []byte

func NewKey(bytes int) *Key {
	key := make(Key, bytes)

	_, err := rand.Read(key)
	if err != nil {
		panic("failed to generate random key")
	}

	return &key
}

// Representation

func (k Key) Base64() string {
	return base64.URLEncoding.EncodeToString(k)
}

// Encryption / Decryption

func (k Key) Encrypt(text string) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(k)
	if err != nil {
		panic(err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext)
}

func (k Key) Decrypt(encryptedText string) string {
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(k)
	if err != nil {
		panic(err)
	}

	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext)
}

// Marshaling

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
