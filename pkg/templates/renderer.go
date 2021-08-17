package templates

import (
	"fmt"
	"html/template"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fiwippi/tanuki/internal/multitemplate"
	"github.com/fiwippi/tanuki/pkg/api/series"
	"github.com/fiwippi/tanuki/pkg/server"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

type Renderer struct {
	multitemplate.Renderer

	server *server.Server
	debug  bool
}

func (r *Renderer) FuncMap() template.FuncMap {
	return template.FuncMap{
		// Versions files so they dont get cached (used when debugging)
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
		"catalog": func() api.Catalog {
			return r.server.Store.GetCatalog()
		},
		// Returns the progress for the user
		"catalogProgress": func(c *gin.Context) map[string]*users.SeriesProgress {
			uid := c.GetString("uid")
			user, err := r.server.Store.GetUser(uid)
			if err != nil {
				return nil
			}
			return user.Progress.Data
		},
		"tags": func() []string {
			return r.server.Store.GetTags().List()
		},
		//
		"seriesWithTag": func(c *gin.Context) api.Catalog {
			tag := c.Param("tag")
			return r.server.Store.GetSeriesWithTag(tag)
		},
		//
		"sid": func(c *gin.Context) string {
			return c.Param("sid")
		},
		//
		"eid": func(c *gin.Context) string {
			return c.Param("eid")
		},
		//
		"entry": func(c *gin.Context) *api.Entry {
			sid := c.Param("sid")
			eid := c.Param("eid")
			e, err := r.server.Store.GetEntry(sid, eid)
			if err != nil {
				return nil
			}
			return e
		},
		//
		"entries": func(c *gin.Context) api.Entries {
			sid := c.Param("sid")
			e, err := r.server.Store.GetEntries(sid)
			if err != nil {
				return nil
			}
			return e
		},
		//
		"entryProgress": func(c *gin.Context) *users.EntryProgress {
			sid := c.Param("sid")
			eid := c.Param("eid")
			uid := c.GetString("uid")

			p, err := series.GetEntryProgressInternal(uid, sid, eid, r.server)
			if err != nil {
				c.Error(err)
				return nil
			}
			return p
		},
		"seriesProgress": func(c *gin.Context) []*users.EntryProgress {
			sid := c.Param("sid")
			uid := c.GetString("uid")

			p, _, err := series.GetSeriesProgressInternal(uid, sid, r.server)
			if err != nil {
				c.Error(err)
				return nil
			}
			return p.Entries
		},
		"series": func(c *gin.Context) api.Series {
			id := c.Param("sid")
			s, err := r.server.Store.GetSeries(id)
			if err != nil {
				return api.Series{}
			}
			return *s
		},
		"missingItems": func() api.MissingItems {
			return r.server.Store.GetMissingItems()
		},
		"username": func(c *gin.Context) string {
			uid := c.GetString("uid")
			u, err := r.server.Store.GetUser(uid)
			if err != nil {
				return ""
			}
			return u.Name
		},
		"users": func() []users.User {
			return r.server.Store.GetUsers(true)
		},
		"user": func(c *gin.Context) users.User {
			uid := c.Query("hash")
			user, err := r.server.Store.GetUser(uid)
			if err != nil {
				return users.User{}
			}
			user.Pass = ""
			return *user
		},
		"mangadexUid": func(c *gin.Context) string {
			return c.Param("uid")
		},
	}
}
