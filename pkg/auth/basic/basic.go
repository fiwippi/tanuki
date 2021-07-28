package basic

import (
	"encoding/base64"
	"errors"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

var (
	ErrAuthHeaderEmpty   = errors.New("auth header is empty")
	ErrInvalidAuthFormat = errors.New("auth header formatted in incorrect way")
)

//
func Auth(realm string, store *bolt.DB) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {
		user, pass, err := parseBasicAuthCred(c)
		if err != nil {
			log.Debug().Err(err).Msg("failed to parse auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		valid, err := store.ValidateLogin(user, pass)
		if !valid || err != nil {
			log.Debug().Err(err).Msg("failed to validate auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
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
