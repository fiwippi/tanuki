package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Middleware which provides ability to log with gin
// using the built-in zerolog logger instead of the
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
		errorMsg := fmt.Errorf("%s", strings.Join(strings.Split(c.Errors.ByType(gin.ErrorTypePrivate).String(), "\n"), ","))

		// User data
		sid := c.Param("sid")
		eid := c.Param("eid")
		uid := c.GetString("uid")
		admin := c.GetBool("admin")

		// Log the data after parsing it
		var event *zerolog.Event
		switch {
		case statusCode >= 400 && statusCode < 500:
			event = Warn().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
				Str("eid", eid).Str("uid", uid).Bool("admin", admin)
		case statusCode >= 500:
			event = Error().Str("method", method).Str("path", path).Str("resp_time", latency).
				Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
				Str("eid", eid).Str("uid", uid).Bool("admin", admin)
		default:
			if strings.HasPrefix(path, "/static") {
				event = Trace().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
					Str("eid", eid).Str("uid", uid).Bool("admin", admin)
			} else {
				event = Info().Str("method", method).Str("path", path).Str("resp_time", latency).
					Int("status", statusCode).Str("client_ip", clientIP).Str("sid", sid).
					Str("eid", eid).Str("uid", uid).Bool("admin", admin)
			}
		}

		if event != nil {
			if len(errorMsg.Error()) > 0 {
				event.Err(errorMsg)
			}
			event.Send()
		}
	}
}
