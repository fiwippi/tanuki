package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/bolt"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

// UserPatchRequest for /api/admin/user
type UserPatchRequest struct {
	NewUsername string     `json:"new_username"`
	NewPassword string     `json:"new_password"`
	NewType     users.Type `json:"new_type"`
}

// UserReply for /api/admin/user
type UserReply struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	User    users.User `json:"user,omitempty"`
}

// PATCH /api/admin/user/:id
func PatchUser(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data UserPatchRequest
		if err := c.ShouldBindJSON(&data); err != nil {
			log.Debug().Str("path", c.Request.URL.Path).Err(err).Msg("")
			c.AbortWithStatusJSON(400, UserReply{Success: false})
			return
		}

		// Ensure user exists
		u, err := s.Store.GetUser(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(500, UserReply{Success: false})
			return
		}

		// Password should be minimum of 8 characters but if the field is left blank then ignored
		if len(data.NewPassword) != 0 {
			if len(data.NewPassword) < 8 {
				c.AbortWithStatusJSON(400, UserReply{Success: false, Message: "password should be minimum of 8 characters"})
				return
			} else if err := s.Store.ChangePassword(u.Hash, data.NewPassword); err != nil {
				c.AbortWithStatusJSON(500, UserReply{Success: false})
				return
			}
		}

		// Change type if needed
		if u.Type != data.NewType {
			if err := s.Store.ChangeUserType(u.Hash, data.NewType); err != nil {
				if err == bolt.ErrNotEnoughAdmins {
					c.AbortWithStatusJSON(500, UserReply{Success: false, Message: "at least one admin must always exist"})
				} else {
					c.AbortWithStatusJSON(500, UserReply{Success: false})
				}
				return
			}
		}

		// Finally change username if needed
		if data.NewUsername != "" && data.NewUsername != u.Name {
			if err := s.Store.ChangeUsername(u.Hash, data.NewUsername); err != nil {
				c.AbortWithStatusJSON(500, UserReply{Success: false})
				return
			}
			// We can ignore the delete error because that only occurs if the
			// cookie is not found and we are setting a new cookie anyway
			s.Session.Delete(c)
			s.Session.Store(hash.SHA1(data.NewUsername), c)
		}

		c.JSON(200, UserReply{Success: true})
	}
}

// DELETE /api/admin/user/:id
func DeleteUser(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if c.GetString("uid") == id { // storde the id as uid not usernameHash
			c.AbortWithStatusJSON(403, UserReply{Success: false, Message: "cannot delete yourself"})
			return
		}

		err := s.Store.DeleteUser(id)
		if err != nil {
			c.AbortWithStatusJSON(500, UserReply{Success: false})
			return
		}

		c.JSON(200, UserReply{Success: true})
	}
}
