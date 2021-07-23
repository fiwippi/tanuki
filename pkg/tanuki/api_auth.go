package tanuki

import (
	"github.com/fiwippi/tanuki/pkg/api"
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

// POST /api/auth/login
func authLogin(c *gin.Context) {
	// Retrieve the request
	var data api.AuthLoginRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		c.AbortWithStatusJSON(400, api.AuthLoginReply{Success: false, Message: ""})
		return
	}

	// Validate login details
	valid, err := db.ValidateLogin(data.Username, data.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed to validate user")
		c.AbortWithStatusJSON(500, api.AuthLoginReply{Success: false})
		return
	} else if !valid {
		c.AbortWithStatusJSON(403, api.AuthLoginReply{Success: false, Message: "invalid username/password"})
		return
	}
	log.Debug().Str("username", data.Username).Msg("validated user")

	// If valid then give user token they can identify themselves with
	usernameHash := auth.SHA1(data.Username)
	session.Store(usernameHash, c)
	c.JSON(200, api.AuthLoginReply{Success: true})
}

// GET /api/auth/logout
func authLogout(c *gin.Context) {
	session.Delete(c)
	uid := c.GetString("uid")
	log.Debug().Str("uid", uid).Msg("user logged out")
	c.JSON(200, api.AuthLogoutReply{Success: true})
}
