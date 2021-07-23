package api

// TagsReply for the route /api/tags
type TagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}

// SeriesTagsRequest for the route /api/tag/:id/series
type SeriesTagsRequest struct {
	Tags []string `json:"tags"`
}

// SeriesTagsReply for the route /api/tag/:id/series
type SeriesTagsReply struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}
