package tanuki

import (
	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/gin-gonic/gin"
)

// These are api routes each user can use to get and
// modify properties about themselves, users are
// expected to provide the cookie identifying themselves
// in order to access/edit their own data unless they're
// an admin

// GET /api/user/type
func apiGetUserType(c *gin.Context) {
	uid := c.GetString("uid")
	u, err := db.GetUser(uid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.UserTypeReply{Success: false})
		return
	}

	c.JSON(200, api.UserTypeReply{Success: true, Type: u.Type})
}

// GET /api/user/name
func apiGetUserName(c *gin.Context) {
	uid := c.GetString("uid")
	u, err := db.GetUser(uid)
	if err != nil {
		c.AbortWithStatusJSON(500, api.UserNameReply{Success: false})
		return
	}

	c.JSON(200, api.UserNameReply{Success: true, Name: u.Name})
}
