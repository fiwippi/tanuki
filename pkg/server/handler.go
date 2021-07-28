package server

import "github.com/gin-gonic/gin"

type HandlerFunc func(s *Server) gin.HandlerFunc
