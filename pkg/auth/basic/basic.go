package basic

import (
	"encoding/base64"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/log"
	"github.com/fiwippi/tanuki/pkg/storage"
)

var (
	ErrAuthHeaderEmpty   = errors.New("auth header is empty")
	ErrInvalidAuthFormat = errors.New("auth header formatted in incorrect way")
)

func Auth(realm string, store *storage.Store) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {
		user, pass, err := parseBasicAuthCred(c)
		if err != nil {
			log.Debug().Err(err).Msg("failed to parse auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(401)
			return
		}

		valid := store.ValidateLogin(user, pass)
		if !valid {
			log.Debug().Err(err).Msg("failed to validate auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(401)
			return
		}

		c.Next()
	}
}

func parseBasicAuthCred(c *gin.Context) (string, string, error) {
	h := c.GetHeader("Authorization")
	if h == "" {
		return "", "", ErrAuthHeaderEmpty
	}

	if !strings.HasPrefix(h, "Basic ") {
		return "", "", ErrInvalidAuthFormat
	}

	encoded := strings.TrimPrefix(h, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(string(decoded), ":")
	if len(parts) != 2 {
		return "", "", ErrInvalidAuthFormat
	}

	return parts[0], parts[1], err
}
