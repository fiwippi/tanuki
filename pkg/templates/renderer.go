package templates

import (
	"html/template"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/multitemplate"
	"github.com/fiwippi/tanuki/pkg/manga"
	"github.com/fiwippi/tanuki/pkg/progress"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/user"
)

type Renderer struct {
	multitemplate.Render

	server *server.Instance
	debug  bool
}

func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		// Whether the user of this context is an admin
		"admin": func(c *gin.Context) bool {
			return c.GetBool("admin")
		},
		// Returns the current catalog (list of all series)
		"catalog": func(c *gin.Context) []manga.Series {
			ctl, err := r.server.Store.GetCatalog()
			if err != nil {
				c.Error(err)
				return []manga.Series{}
			}
			return ctl
		},
		// Returns the progress for the user
		"catalogProgress": func(c *gin.Context) progress.Catalog {
			uid := c.GetString("uid")
			cp, err := r.server.Store.GetCatalogProgress(uid)
			if err != nil {
				c.Error(err)
				return progress.Catalog{}
			}
			return cp
		},
		"tags": func(c *gin.Context) []string {
			t, err := r.server.Store.GetAllTags()
			if err != nil {
				c.Error(err)
				return nil
			}
			return t.List()
		},
		"seriesWithTag": func(c *gin.Context) []manga.Series {
			tag := c.Param("tag")
			series, err := r.server.Store.GetSeriesWithTag(tag)
			if err != nil {
				c.Error(err)
				return nil
			}
			return series
		},
		"sid": func(c *gin.Context) string {
			return c.Param("sid")
		},
		"eid": func(c *gin.Context) string {
			return c.Param("eid")
		},
		"entry": func(c *gin.Context) manga.Entry {
			sid := c.Param("sid")
			eid := c.Param("eid")
			e, err := r.server.Store.GetEntry(sid, eid)
			if err != nil {
				c.Error(err)
				return manga.Entry{}
			}
			return e
		},
		"entries": func(c *gin.Context) []manga.Entry {
			sid := c.Param("sid")
			e, err := r.server.Store.GetEntries(sid)
			if err != nil {
				c.Error(err)
				return []manga.Entry{}
			}
			return e
		},
		"entryProgress": func(c *gin.Context) progress.Entry {
			sid := c.Param("sid")
			eid := c.Param("eid")
			uid := c.GetString("uid")

			ep, err := r.server.Store.GetEntryProgress(sid, eid, uid)
			if err != nil {
				c.Error(err)
				return progress.Entry{}
			}
			return ep
		},
		"seriesProgress": func(c *gin.Context) progress.Series {
			sid := c.Param("sid")
			uid := c.GetString("uid")

			p, err := r.server.Store.GetSeriesProgress(sid, uid)
			if err != nil {
				c.Error(err)
				return progress.Series{}
			}
			return p
		},
		"series": func(c *gin.Context) manga.Series {
			id := c.Param("sid")
			s, err := r.server.Store.GetSeries(id)
			if err != nil {
				return manga.Series{}
			}
			return s
		},
		"username": func(c *gin.Context) string {
			uid := c.GetString("uid")
			u, err := r.server.Store.GetUser(uid)
			if err != nil {
				c.Error(err)
				return ""
			}
			return u.Name
		},
		"users": func(c *gin.Context) []user.Account {
			u, err := r.server.Store.GetUsers()
			if err != nil {
				c.Error(err)
				return []user.Account{}
			}
			for i := range u {
				u[i].Pass = ""
			}
			return u
		},
		"user": func(c *gin.Context) user.Account {
			uid := c.Query("uid")
			u, err := r.server.Store.GetUser(uid)
			if err != nil {
				return user.Account{}
			}
			u.Pass = ""
			return u
		},
		"mangadexUUID": func(c *gin.Context) string {
			return c.Param("uuid")
		},
	}
}
