package api

import "github.com/fiwippi/tanuki/pkg/core"

// Admin routes
// GET, PATCH, DELETE /api/admin/user/:id
// GET, PUT /api/admin/users
// GET /api/admin/db

type AdminUserPatchRequest struct {
	NewUsername string        `json:"new_username"`
	NewPassword string        `json:"new_password"`
	NewType     core.UserType `json:"new_type"`
}

// AdminUserReply for /api/admin/user
type AdminUserReply struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	User    core.User `json:"user,omitempty"`
}

type AdminUsersPutRequest struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Type     core.UserType `json:"type"`
}

// AdminUsersReply for /api/admin/users
type AdminUsersReply struct {
	Success bool        `json:"success"`
	Users   []core.User `json:"users,omitempty"`
	Message string      `json:"message,omitempty"`
}

// AdminDBReply for /api/admin/db
type AdminDBReply struct {
	Success bool   `json:"success"`
	DB      string `json:"db"`
}

// AdminLibraryReply for /api/admin/library
type AdminLibraryReply struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// AdminLibraryMissingEntriesReply for /api/admin/library/missing-entries
type AdminLibraryMissingEntriesReply struct {
	Success bool         `json:"success"`
	Entries MissingItems `json:"entries"`
}
