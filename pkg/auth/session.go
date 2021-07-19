package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/shaj13/libcache"
	_ "github.com/shaj13/libcache/lru"
)

// Session stores a value in a cookie and encrypts it using a SecureKey.
// A Time To Live (TTL) is specified, this is the duration the client
// should hold onto the cookie.
type Session struct {
	cache      libcache.Cache
	secret     SecureKey
	cookieName string
	ttl        time.Duration
}

func NewSession(ttl time.Duration, cookie string, secret SecureKey) *Session {
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
	return Decrypt(encryptedID, s.secret), nil
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
	key := xid.New().String()
	s.cache.Store(key, value)

	// Store encrypted version in the cookie store
	encryptedID := Encrypt(key, s.secret)
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
	decryptedID := Decrypt(encryptedID, s.secret)
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

