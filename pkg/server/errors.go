package server

import "github.com/gin-gonic/gin"

func (s *Server) SetErr404(f gin.HandlerFunc) {
	s.err404 = f
	s.Router.NoRoute(f)
}

func (s *Server) Err404(c *gin.Context) {
	if s.err404 != nil {
		s.err404(c)
		return
	}

	panic("server has no 404 function")
}
