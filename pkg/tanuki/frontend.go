package tanuki

import (
	"github.com/gin-gonic/gin"
)

// GET /
func home(c *gin.Context) {
	c.HTML(200, "home.tmpl", nil)
}

// GET /entries/:sid
func entries(c *gin.Context) {
	id := c.Param("sid")
	_, err := db.GetSeries(id)
	if err != nil {
		err404(c)
		return
	}
	c.HTML(200, "entries.tmpl", nil)
}

// GET /tags/:tag
func specificTag(c *gin.Context) {
	tag := c.Param("tag")

	allTags := db.GetTags()
	if !allTags.Has(tag) {
		err404(c)
		return
	}
	c.HTML(200, "specific-tag.tmpl", nil)
}

// GET /tags
func tags(c *gin.Context) {
	c.HTML(200, "tags.tmpl", nil)
}

// GET /reader/:sid/:eid
func reader(c *gin.Context) {
	sid := c.Param("sid")
	eid := c.Param("eid")

	_, err := db.GetEntry(sid, eid)
	if err != nil {
		err404(c)
		return
	}

	c.HTML(200, "reader.tmpl", nil)
}