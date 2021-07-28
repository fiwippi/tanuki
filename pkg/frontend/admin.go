package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

// GET /admin
func adminDashboard(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin.tmpl", nil)
	}
}

// GET /admin/missing-items
func adminMissingItems(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-missing-items.tmpl", nil)

	}
}

// GET /admin/users
func adminUsers(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users.tmpl", nil)
	}
}

// GET /admin/users/edit
func adminUsersEdit(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-edit.tmpl", nil)
	}
}

// GET /admin/users/create
func adminUsersCreate(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "admin-users-create.tmpl", nil)
	}
}