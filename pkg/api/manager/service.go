package manager

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/auth/cookie"
	"github.com/fiwippi/tanuki/pkg/server"
)

func dlFunc(fn func()) server.HandlerFunc {
	return func(_ *server.Instance) gin.HandlerFunc {
		return func(c *gin.Context) {
			fn()
			c.Status(200)
		}
	}
}

func dlFuncErr(fn func() error) server.HandlerFunc {
	return func(_ *server.Instance) gin.HandlerFunc {
		return func(c *gin.Context) {
			if err := fn(); err != nil {
				c.AbortWithError(500, err)
				return
			}
			c.Status(200)
		}
	}
}

func NewService(g *server.RouterGroup) {
	m := g.Group("/manager")
	m.Use(cookie.Auth(g.Server, cookie.Abort))

	dl := m.Group("/download")
	dl.GET("/", viewDownloads)
	dl.POST("/chapters", downloadChapters)
	dl.GET("/delete-all-dl", dlFuncErr(g.Server.Manager.DeleteAllDownloads))
	dl.GET("/delete-successful-dl", dlFuncErr(g.Server.Manager.DeleteSuccessfulDownloads))
	dl.GET("/retry-failed-dl", dlFuncErr(g.Server.Manager.RetryFailedDownloads))
	dl.GET("/pause-dl", dlFunc(g.Server.Manager.Pause))
	dl.GET("/resume-dl", dlFunc(g.Server.Manager.Resume))
	dl.GET("/cancel-dl", dlFunc(g.Server.Manager.CancelDownloads))
}
