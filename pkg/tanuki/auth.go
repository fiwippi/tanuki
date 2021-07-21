package tanuki

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/auth"
)

const authTime = time.Minute * 30 // How long should a session last for
const authCookieName = "tanuki"   // Name of the cookie stored on the client

var session *auth.Session

// We store the username hash in the session, it become more
// efficient to access the database and improve security

// GET /login
func login(c *gin.Context) {
	c.HTML(200, "login.tmpl", nil)
}

// POST /auth/login
func authLogin(c *gin.Context) {
	// Retrieve the request
	var data auth.LoginRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.AbortWithStatusJSON(400, auth.LoginReply{Success: false, Message: ""})
		return
	}

	// Validate login details
	valid, err := db.ValidateLogin(data.Username, data.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate user")
		c.AbortWithStatusJSON(500, auth.LoginReply{Success: false})
		return
	} else if !valid {
		c.AbortWithStatusJSON(403, auth.LoginReply{Success: false, Message: "invalid username/password"})
		return
	}
	log.Debug().Str("username", data.Username).Msg("validated user")

	// If valid then give user token they can identify themselves with
	usernameHash := auth.HashSHA1(data.Username)
	session.Store(usernameHash, c)
	c.JSON(200, auth.LoginReply{Success: true})
}

// GET /auth/logout
func authLogout(c *gin.Context) {
	session.Delete(c)
	uid := c.GetString("uid")
	log.Debug().Str("uid", uid).Msg("user logged out")
	c.JSON(200, auth.LogoutReply{Success: true})
}
