package auth

import (
	"errors"
	"github.com/fiwippi/tanuki/internal/encryption"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie")
	ErrNotInCache    = errors.New("item not found in cache")
)

// Session stores a value in a cookie and encrypts it using a SecureKey.
// A Time To Live (TTL) is specified, this is the duration the client
// should hold onto the cookie.
type Session struct {
	cache      libcache.Cache // Locally store the data
	secret     encryption.Key // Used to hash the uid in the cookie
	cookieName string         // What name should the cookie have
	ttl        time.Duration  // Time to live of the cookie and the key in the cache
}

func NewSession(ttl time.Duration, cookie string, secret encryption.Key) *Session {
	s := &Session{
		cache:      libcache.LRU.New(0),
		secret:     secret,
		cookieName: cookie,
		ttl:        ttl,
	}
	s.cache.SetTTL(ttl)
	s.cache.RegisterOnExpired(func(key, _ interface{}) {
		s.cache.Delete(key)
	})

	return s
}

func (s *Session) getDecryptedID(c *gin.Context) (string, error) {
	// Get the encrypted user key
	encryptedID, err := c.Cookie(s.cookieName)
	if err != nil {
		return "", err
	}

	// Cookie found, decrypt it to get the key
	return s.secret.Decrypt(encryptedID), nil
}

func (s *Session) getValue(c *gin.Context) (string, error) {
	// Cookie found, decrypt it to get the key
	decryptedID, err := s.getDecryptedID(c)
	if err != nil {
		return "", err
	}

	value, found := s.cache.Load(decryptedID)
	if !found {
		// Cookie invalid
		return "", ErrInvalidCookie
	}

	// Return the cookies value
	return value.(string), nil
}

func (s *Session) TTL() int {
	return int(s.ttl.Seconds())
}

func (s *Session) Store(value string, c *gin.Context) {
	// Store unencrypted version in local cache
	key := encryption.NewKey(12).Base64()
	s.cache.Store(key, value)

	// Store encrypted version in the cookie store
	encryptedID := s.secret.Encrypt(key)
	c.SetCookie(s.cookieName, encryptedID, s.TTL(), "/", "", false, true)
}

func (s *Session) TimeLeft(c *gin.Context) (time.Duration, error) {
	// Get the encrypted user key
	decryptedID, err := s.getDecryptedID(c)
	if err != nil {
		return 0, err
	}

	// Get the expiry time
	t, found := s.cache.Expiry(decryptedID)
	if !found {
		return 0, ErrNotInCache
	}
	return t.Sub(time.Now()), nil
}

func (s *Session) Refresh(c *gin.Context) error {
	encryptedID, err := c.Cookie(s.cookieName)
	if err != nil {
		return err
	}
	decryptedID := s.secret.Decrypt(encryptedID)
	value, found := s.cache.Load(decryptedID)
	if !found {
		return ErrInvalidCookie
	}

	// Refresh the cookie
	s.cache.StoreWithTTL(decryptedID, value, s.ttl)
	c.SetCookie(s.cookieName, encryptedID, s.TTL(), "/", "", false, true)
	return nil
}

func (s *Session) Delete(c *gin.Context) {
	decryptedID, err := s.getDecryptedID(c)
	if err != nil {
		// Err means cookie not found, so it's
		// already been deleted which is fine
		return
	}

	s.cache.Delete(decryptedID)
}

func (s *Session) Get(c *gin.Context) (string, error) {
	return s.getValue(c)
}
