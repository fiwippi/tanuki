package tanuki

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/pkg/api"
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
	database "github.com/fiwippi/tanuki/pkg/db"
)

// GET /api/admin/users
func apiGetAdminUsers(c *gin.Context) {
	users, err := db.GetUsers(true)
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminUsersReply{Success: false, Users: nil})
		return
	}

	c.JSON(200, api.AdminUsersReply{Success: true, Users: users})
}

// PUT /api/admin/users
func apiPutAdminUsers(c *gin.Context) {
	var data api.AdminUsersPutRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.AdminUserReply{Success: false})
		return
	}

	if data.Username == "" {
		c.AbortWithStatusJSON(400, api.AdminUserReply{Success: false, Message: "username cannot be empty"})
		return
	}

	if len(data.Password) < 8 {
		c.AbortWithStatusJSON(400, api.AdminUserReply{Success: false, Message: "password should be minimum of 8 characters"})
		return
	}

	err := db.CreateUser(core.NewUser(data.Username, data.Password, data.Type))
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
		return
	}

	c.JSON(200, api.AdminUserReply{Success: true})
}

// GET /api/admin/user/:id
func apiGetAdminUser(c *gin.Context) {
	user, err := db.GetUser(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
		return
	}
	user.Pass = ""

	c.JSON(200, api.AdminUserReply{Success: true, User: *user})
}

// PATCH /api/admin/user/:id
func apiPatchAdminUser(c *gin.Context) {
	var data api.AdminUserPatchRequest
	if err := c.ShouldBindJSON(&data); err != nil {
		log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
		c.AbortWithStatusJSON(400, api.AdminUserReply{Success: false})
		return
	}

	// Ensure user exists
	u, err := db.GetUser(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
		return
	}

	// Password should be minimum of 8 characters but if the field is left blank then ignored
	if len(data.NewPassword) != 0 {
		if len(data.NewPassword) < 8 {
			c.AbortWithStatusJSON(400, api.AdminUserReply{Success: false, Message: "password should be minimum of 8 characters"})
			return
		} else if err := db.ChangePassword(u.Hash, data.NewPassword); err != nil {
			c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
			return
		}
	}

	// Change type if needed
	if u.Type != data.NewType {
		if err := db.ChangeUserType(u.Hash, data.NewType); err != nil {
			if err == database.ErrNotEnoughAdmins {
				c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false, Message: "at least one admin must always exist"})
			} else {
				c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
			}
			return
		}
	}

	// Finally change username if needed
	if data.NewUsername != "" && data.NewUsername != u.Name {
		if err := db.ChangeUsername(u.Hash, data.NewUsername); err != nil {
			c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
			return
		}
		// We don't ignore about the delete error because that only occurs if
		// the cookie is not found but we can ignore that since we are setting
		// a new cookie anyway
		session.Delete(c)
		session.Store(auth.HashSHA1(data.NewUsername), c)
	}

	c.JSON(200, api.AdminUserReply{Success: true})
}

// DELETE /api/admin/user/:id
func apiDeleteAdminUser(c *gin.Context) {
	id := c.Param("id")
	if c.GetString("uid") == id { // storde the id as uid not usernameHash
		c.AbortWithStatusJSON(403, api.AdminUserReply{Success: false, Message: "cannot delete yourself"})
		return
	}

	err := db.DeleteUser(id)
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminUserReply{Success: false})
		return
	}

	c.JSON(200, api.AdminUserReply{Success: true})
}

// GET /api/admin/db
func apiGetAdminDB(c *gin.Context) {
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=\"db.txt\"")
	c.Data(200, "text/plain", []byte(db.String()))
}

// GET /api/admin/library/scan
func apiGetAdminLibraryScan(c *gin.Context) {
	now := time.Now()
	err := ScanLibrary()
	if err != nil {
		log.Error().Err(err).Msg("failed to scan library")
		c.AbortWithStatusJSON(500, api.AdminLibraryReply{Success: false})
	} else {
		timeTaken := time.Now().Sub(now)
		c.JSON(200, api.AdminLibraryReply{Success: true, Message: fmt.Sprintf("The time taken was %s", timeTaken)})
	}
}

// GET /api/admin/library/generate-thumbnails
func apiGetAdminLibraryGenerateThumbnails(c *gin.Context) {
	now := time.Now()
	err := db.GenerateThumbnails(true)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate thumbnails")
		c.AbortWithStatusJSON(500, api.AdminLibraryReply{Success: false})
	} else {
		timeTaken := time.Now().Sub(now)
		c.JSON(200, api.AdminLibraryReply{Success: true, Message: fmt.Sprintf("The time taken was %s", timeTaken)})
	}
}

// GET /api/admin/library/missing-entries
func apiGetAdminLibraryMissingEntries(c *gin.Context) {
	e := db.GetMissingEntries()
	c.JSON(200, api.AdminLibraryMissingEntriesReply{Success: true, Entries: e})
}

// DELETE /api/admin/library/missing-entries
func apiDeleteAdminLibraryMissingEntries(c *gin.Context) {
	err := db.DeleteMissingEntries()
	if err != nil {
		c.AbortWithStatusJSON(500, api.AdminLibraryMissingEntriesReply{Success: false})
		return
	}

	c.JSON(200, api.AdminLibraryMissingEntriesReply{Success: true})
}
