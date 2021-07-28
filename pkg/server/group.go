package server

import "github.com/gin-gonic/gin"

type RouterGroup struct {
	*gin.RouterGroup

	Server *Server
}

func (rg *RouterGroup) Group(relativePath string) *RouterGroup {
	return &RouterGroup{
		RouterGroup: rg.RouterGroup.Group(relativePath),
		Server:      rg.Server,
	}
}

func (rg *RouterGroup) Use(middleware ...gin.HandlerFunc) *RouterGroup {
	newRg := rg.RouterGroup.Use(middleware...).(*gin.RouterGroup)

	return &RouterGroup{
		RouterGroup: newRg,
		Server:      rg.Server,
	}
}

func (rg *RouterGroup) GET(relativePath string, f HandlerFunc) {
	rg.RouterGroup.GET(relativePath, f(rg.Server))
}

func (rg *RouterGroup) POST(relativePath string, f HandlerFunc) {
	rg.RouterGroup.POST(relativePath, f(rg.Server))
}

func (rg *RouterGroup) DELETE(relativePath string, f HandlerFunc) {
	rg.RouterGroup.DELETE(relativePath, f(rg.Server))
}

func (rg *RouterGroup) PATCH(relativePath string, f HandlerFunc) {
	rg.RouterGroup.PATCH(relativePath, f(rg.Server))
}

func (rg *RouterGroup) PUT(relativePath string, f HandlerFunc) {
	rg.RouterGroup.PUT(relativePath, f(rg.Server))
}
