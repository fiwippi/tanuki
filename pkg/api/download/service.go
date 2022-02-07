package download

import (
	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/downloading"
	"github.com/fiwippi/tanuki/pkg/server"
)

func NewService(api *server.RouterGroup) {
	// Setup the download manager
	manager = downloading.NewManager(api.Server.Mangadex, api.Server.Conf.Paths.Library, api.Server.Store, 5)

	// Setup routes
	g := api.Group("/download")
	g.Use(cookie.Auth(api.Server, cookie.Abort))

	g.POST("/chapters", GetChapters)
	g.GET("/manager", ViewManager)
	g.GET("/manager/delete-finished-tasks", DeleteFinishedTasks)
	g.GET("/manager/retry-failed-tasks", RetryFailedTasks)
	g.GET("/manager/pause-downloads", PauseDownloads)
	g.GET("/manager/resume-downloads", ResumeDownloads)
	g.GET("/manager/cancel-downloads", CancelDownloads)
}
