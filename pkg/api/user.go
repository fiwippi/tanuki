package api

import "github.com/fiwippi/tanuki/pkg/core"

// User routes, the cookie is used to identify
// the user in this scenario as opposed to using
// the user id
// GET /api/user/type
// GET /api/user/name
// GET, PATCH api/user/progress?series=xxx&manga=xxx

type UserProgressRequest struct {
	Progress string `json:"progress"`
}

// UserPropertyReply defines the reply from /api/user/:property
type UserPropertyReply struct {
	Success         bool          `json:"success"`
	Type            core.UserType `json:"type,omitempty"`
	Username        string        `json:"username,omitempty"`
	ProgressPercent float64       `json:"progress_percent"`
	ProgressPage    int           `json:"progress_page"`
}
