package logging

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Middleware which provides ability to log with gin
// using the built in zerolog logger instead of the
// default one
func Middleware() gin.HandlerFunc {
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
		errorMsg := strings.Join(strings.Split(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"), ",")

		// User data
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")
		admin := c.GetBool("admin")

		// Log the data after parsing it
		switch {
		case statusCode >= 400 && statusCode < 500:
			log.Warn().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
				Str("eid", eid).Str("uid", uid).Bool("admin", admin).Msg(errorMsg)
		case statusCode >= 500:
			log.Error().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
				Str("eid", eid).Str("uid", uid).Bool("admin", admin).Msg(errorMsg)
		default:
			if strings.HasPrefix(path, "/static") {
				log.Trace().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
					Str("eid", eid).Str("uid", uid).Bool("admin", admin).Msg(errorMsg)
			} else {
				log.Info().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
					Str("eid", eid).Str("uid", uid).Bool("admin", admin).Msg(errorMsg)
			}
		}
	}
}
