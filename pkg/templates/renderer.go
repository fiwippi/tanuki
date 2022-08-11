package templates

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/platform/multitemplate"
	"github.com/fiwippi/tanuki/pkg/human"
	"github.com/fiwippi/tanuki/pkg/manga"

	"github.com/fiwippi/tanuki/pkg/server"
)

type Renderer struct {
	multitemplate.Renderer

	server *server.Instance
	debug  bool
}

func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		// Versions files so they don't get cached (used when debugging)
		"versioning": func(filePath string) string {
			if r.debug {
				return fmt.Sprintf("%s?q=%s", filePath, strconv.Itoa(int(time.Now().Unix())))
			}
			return filePath
		},
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
		"catalogProgress": func(c *gin.Context) human.CatalogProgress {
			uid := c.GetString("uid")
			cp, err := r.server.Store.GetCatalogProgress(uid)
			if err != nil {
				c.Error(err)
				return human.CatalogProgress{}
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
		"entryProgress": func(c *gin.Context) human.EntryProgress {
			sid := c.Param("sid")
			eid := c.Param("eid")
			uid := c.GetString("uid")

			ep, err := r.server.Store.GetEntryProgress(sid, eid, uid)
			if err != nil {
				c.Error(err)
				return human.EntryProgress{}
			}
			return ep
		},
		"seriesProgress": func(c *gin.Context) human.SeriesProgress {
			sid := c.Param("sid")
			uid := c.GetString("uid")

			p, err := r.server.Store.GetSeriesProgress(sid, uid)
			if err != nil {
				c.Error(err)
				return human.SeriesProgress{}
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
		"users": func(c *gin.Context) []human.User {
			u, err := r.server.Store.GetUsers()
			if err != nil {
				c.Error(err)
				return []human.User{}
			}
			for i := range u {
				u[i].Pass = ""
			}
			return u
		},
		"user": func(c *gin.Context) human.User {
			uid := c.Query("uid")
			user, err := r.server.Store.GetUser(uid)
			if err != nil {
				return human.User{}
			}
			user.Pass = ""
			return user
		},
		"mangadexUUID": func(c *gin.Context) string {
			return c.Param("uuid")
		},
		"subscriptions": func(c *gin.Context) []manga.Subscription {
			sub, err := r.server.Store.GetAllSubscriptions()
			if err != nil {
				c.AbortWithError(500, err)
				return nil
			}
			return sub
		},
	}
}
