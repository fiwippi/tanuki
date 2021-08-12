package frontend

import (
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/gin-gonic/gin"
)

func downloadHome(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download.tmpl", c)
	}
}

func downloadMangadex(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "search-mangadex.tmpl", c)
	}
}

func downloadMangadexChapters(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download-mangadex-chapters.tmpl", c)
	}
}

func downloadManager(s *server.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(200, "download-manager.tmpl", c)
	}
}
