package cookie

import (
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"time"
)

// Auth middleware which ensures the user is authorised
func Auth(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := s.Session.Get(c)
		if err != nil {
			// Invalid cookie
			log.Debug().Err(err).Msg("failed to auth request")
			c.Redirect(302, "/login")
			return
		}

		// Set values for next requests
		valid, _ := s.Store.IsAdmin(uid)
		c.Set("admin", valid)
		c.Set("uid", uid)

		// Refresh cookie for the user, only refresh if the cookie is bout to expire
		timeLeft, err := s.Session.TimeLeft(c)
		if err != nil {
			log.Debug().Err(err).Str("uid", uid).Msg("failed to get time left")
		}
		if err == nil && timeLeft < (time.Minute*3) {
			err = s.Session.Refresh(c)
			if err != nil {
				log.Debug().Err(err).Str("uid", uid).Msg("failed to refresh cookie")
			}
		}

		// Cookie valid
		c.Next()
	}
}

// SkipIfAuthed Middleware which skips these routes if the user is
// already authorised and redirect to the home page instead
func SkipIfAuthed(session *auth.Session, home string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := session.Get(c)
		if err == nil {
			c.Redirect(302, home)
			return
		}

		c.Next()
	}
}

// Admin Middleware which ensures the user accessing the route
// must be an admin, should be called after CookieEnsureAuthed
func Admin(home string) gin.HandlerFunc { // ||
	return func(c *gin.Context) {
		admin := c.GetBool("admin")
		if !admin {
			c.Redirect(302, home)
			return
		}

		c.Next()
	}
}
