package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func adminDashboard(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin.tmpl", c)
	}
}

func adminMissingItems(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-missing-items.tmpl", c)

	}
}

func adminUsers(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users.tmpl", c)
	}
}

func adminUsersEdit(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-edit.tmpl", c)
	}
}

func adminUsersCreate(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-create.tmpl", c)
	}
}
