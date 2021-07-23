package tanuki

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Middleware which provides ability to log with gin
// using the built in zerolog logger instead of the
// default one
func logMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Before Request
		t := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// After request
		latency := time.Since(t).String()
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		if raw != "" {
			path = path + "?" + raw
		}
		errorMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Log the data after parsing it
		switch {
		case statusCode >= 400 && statusCode < 500:
			log.Warn().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Msg(errorMsg)
		case statusCode >= 500:
			log.Error().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Msg(errorMsg)
		default:
			if strings.HasPrefix(path, "/static") {
				log.Trace().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Msg(errorMsg)
			} else {
				log.Info().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Msg(errorMsg)
			}
		}
	}
}

// Middleware which skips these routes if the user is
// already authorised and redirect to the home page instead
func skipIfAuthedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := session.Get(c)
		if err == nil {
			c.Redirect(302, "/")
			return
		}

		c.Next()
	}
}

// Middleware which ensures the user is authorised
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := session.Get(c)
		if err != nil {
			// Invalid cookie
			log.Debug().Err(err).Msg("failed to auth request")
			c.Redirect(302, "/login")
			return
		}

		// Set values for next requests
		valid, _ := db.IsAdmin(uid)
		c.Set("admin", valid)
		c.Set("uid", uid)

		// Refresh cookie for the user, only refresh if the cookie is bout to expire
		timeLeft, err := session.TimeLeft(c)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Msg("failed to get time left")
		}
		if err == nil && timeLeft < (time.Minute*3) {
			err = session.Refresh(c)
			if err != nil {
				log.Debug().Err(err).Str("uid", uid).Msg("failed to refresh cookie")
			}
		}

		// Cookie valid
		c.Next()
	}
}

// Middleware which ensures the user accessing the route
// must be an admin, should be called after authMiddleware
func adminMiddleware() gin.HandlerFunc { // ||
	return func(c *gin.Context) {
		admin := c.GetBool("admin")
		if !admin {
			c.Redirect(302, "/")
			return
		}

		c.Next()
	}
}

//
func basicAuthMiddleware(realm string) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {
		user, pass, err := basicAuthCredentials(c)
		if err != nil {
			log.Debug().Err(err).Msg("failed to parse auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		valid, err := db.ValidateLogin(user, pass)
		if !valid || err != nil {
			log.Debug().Err(err).Msg("failed to parse auth credentials")
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}
}

func basicAuthCredentials(c *gin.Context) (string, string, error) {
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
