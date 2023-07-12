package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/hash"

	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/storage"
	"github.com/fiwippi/tanuki/pkg/user"
)

func PatchUser(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := struct {
			NewUsername string    `json:"new_username"`
			NewPassword string    `json:"new_password"`
			NewType     user.Type `json:"new_type"`
		}{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.AbortWithError(400, err)
			return
		}

		// Ensure user exists
		u, err := s.Store.GetUser(c.Param("id"))
		if err != nil {
			c.AbortWithStatus(400)
			return
		}

		// Password should be minimum of 8 characters but if the field is left blank then ignored
		if len(data.NewPassword) != 0 {
			if len(data.NewPassword) < 8 {
				c.AbortWithStatusJSON(400, gin.H{"message": "password should be minimum of 8 characters"})
				return
			} else if err := s.Store.ChangePassword(u.UID, data.NewPassword); err != nil {
				c.AbortWithError(500, err)
				return
			}
		}

		// Change type if needed
		if u.Type != data.NewType {
			if err := s.Store.ChangeType(u.UID, data.NewType); err != nil {
				if err == storage.ErrNotEnoughAdmins {
					c.AbortWithStatusJSON(400, gin.H{"message": "at least one admin must always exist"})
				} else {
					c.AbortWithError(500, err)
				}
				return
			}
		}

		// Finally change username if needed
		if data.NewUsername != "" && data.NewUsername != u.Name {
			if err := s.Store.ChangeUsername(u.UID, data.NewUsername); err != nil {
				c.AbortWithError(500, err)
				return
			}
			// Refresh the session cookie
			s.Session.Delete(c)
			s.Session.Store(hash.SHA1(data.NewUsername), c)
		}

		c.Status(200)
	}
}

func DeleteUser(s *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("id")
		if c.GetString("uid") == uid {
			c.AbortWithStatusJSON(403, gin.H{"message": "cannot delete yourself"})
			return
		}

		err := s.Store.DeleteUser(uid)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.Status(200)
	}
}
