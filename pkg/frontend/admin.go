package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func adminDashboard(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin.tmpl", c)
	}
}

func adminMissingItems(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-missing-items.tmpl", c)

	}
}

func adminUsers(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users.tmpl", c)
	}
}

func adminUsersEdit(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-edit.tmpl", c)
	}
}

func adminUsersCreate(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-create.tmpl", c)
	}
}
