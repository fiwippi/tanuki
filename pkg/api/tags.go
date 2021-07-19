package api

// GET /api/tags
// GET /api/tag/:id/series

type AllTagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}
