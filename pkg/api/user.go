package api

import "github.com/fiwippi/tanuki/pkg/core"

// User routes, the cookie is used to identify
// the user in this scenario as opposed to using
// the user id
// GET /api/user/type
// GET /api/user/name

// UserTypeReply defines the reply from /api/user/type
type UserTypeReply struct {
	Success bool          `json:"success"`
	Type    core.UserType `json:"type"`
}

// UserNameReply defines the reply from /api/user/name
type UserNameReply struct {
	Success bool   `json:"success"`
	Name    string `json:"name"`
}
