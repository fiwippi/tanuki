package tanuki

import (
	"github.com/gin-gonic/gin"

	"errors"
)

var (
	ErrAuthHeaderEmpty = errors.New("auth header is empty")
	ErrInvalidAuthFormat = errors.New("auth header formatted in incorrect way")
)

func err404(c *gin.Context) {
	c.HTML(404, "404.tmpl", nil)
}

func err500(c *gin.Context) {
	c.HTML(500, "500.tmpl", nil)
}
