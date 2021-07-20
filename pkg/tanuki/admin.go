package tanuki

import (
	"github.com/gin-gonic/gin"
)

// GET /admin
func adminDashboard(c *gin.Context) {
	c.HTML(200, "admin.tmpl", nil)
}

// GET /admin/missing-entries
func adminMissingEntries(c *gin.Context) {
	c.HTML(200, "admin-missing-entries.tmpl", nil)
}

// GET /admin/users
func adminUsers(c *gin.Context) {
	c.HTML(200, "admin-users.tmpl", nil)
}

// GET /admin/users/edit
func adminUsersEdit(c *gin.Context) {
	c.HTML(200, "admin-users-edit.tmpl", nil)
}

// GET /admin/users/create
func adminUsersCreate(c *gin.Context) {
	c.HTML(200, "admin-users-create.tmpl", nil)
}
