package auth

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"hash"
)

// SHA256 hashes a string using SHA256
func SHA256(s string) string {
	return hashStr(s, sha256.New())
}

// SHA1 hashes a string using SHA1
func SHA1(s string) string {
	return hashStr(s, sha1.New())
}

func hashStr(s string, h hash.Hash) string {
	h.Write([]byte(s))
	digest := h.Sum(nil)
	return base64.URLEncoding.EncodeToString(digest)
}
