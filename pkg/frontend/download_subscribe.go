package frontend

import (
	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/pkg/server"
)

func downloadSubscribeHome(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download-subscribe.tmpl", c)
	}
}

func downloadMangadex(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "search-mangadex.tmpl", c)
	}
}

func downloadMangadexChapters(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download-mangadex-chapters.tmpl", c)
	}
}

func downloadManager(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download-manager.tmpl", c)
	}
}

func subscriptionManager(_ *server.Instance) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "subscription-manager.tmpl", c)
	}
}
