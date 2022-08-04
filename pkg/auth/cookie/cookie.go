package cookie

import (
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/server"
)

var RedirectCookie = "tanuki-redirect"

// Auth middleware which ensures the user is authorised
func Auth(s *server.Instance, action Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, err := s.Session.Get(c)
		if err != nil {
			// Invalid cookie
			switch action {
			case Redirect:
				createRedirectCookie(c)
				c.Redirect(302, "/login")
			case Abort:
				c.AbortWithError(401, err)
			}
			return
		}

		// Set values for next requests
		valid := s.Store.IsAdmin(uid)
		c.Set("admin", valid)
		c.Set("uid", uid)

		// Refresh cookie for the user
		err = s.Session.Refresh(c)
		if err != nil {
			c.Error(err)
		}

		// Cookie valid
		uRaw, err := c.Cookie(RedirectCookie)
		if err != nil || uRaw == "" {
			c.Next()
		} else {
			// If we are accessing a reader route we remove the
			// ?page= query since that will reset the progress
			// the user is on
			u, err := url.Parse(uRaw)
			if err != nil {
				c.Next()
				return
			}
			u.RawQuery = ""
			deleteRedirectCookie(c)
			c.Redirect(302, u.String())
		}
	}
}

// SkipIfAuthed Middleware which skips these routes if the user is
// already authorised and redirect to the home page instead
func SkipIfAuthed(session *auth.Session, home string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := session.Get(c)
		if err == nil {
			createRedirectCookie(c)
			c.Redirect(302, home)
			c.Abort()
			return
		}

		c.Next()
	}
}

// Admin Middleware which ensures the user accessing the route
// must be an admin, should be called after CookieEnsureAuthed
func Admin(home string) gin.HandlerFunc {
	return func(c *gin.Context) {
		admin := c.GetBool("admin")
		if !admin {
			c.Redirect(302, home)
			c.Abort()
			return
		}

		c.Next()
	}
}

func createRedirectCookie(c *gin.Context) {
	c.SetCookie(RedirectCookie, c.Request.URL.String(), 5*60*60, "/", "", false, true)
}

func deleteRedirectCookie(c *gin.Context) {
	c.SetCookie(RedirectCookie, "", 0, "/", "", false, true)
}
