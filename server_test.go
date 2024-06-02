package tanuki

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// It's cleaner to write tests
	// which expect output that's
	// indented
	xmlIndent = "  "
}

// Server Config

func TestServerConfig_JSON(t *testing.T) {
	c1 := DefaultServerConfig()
	data, err := json.MarshalIndent(c1, "", " ")
	require.NoError(t, err)
	fmt.Println(string(data))

	var c2 ServerConfig
	require.NoError(t, json.Unmarshal(data, &c2))
	require.Equal(t, c1, c2)
}

// Server

func TestServer_GetSearch(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)
	r := router(s)

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/opds/v1.2/search", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid auth", func(t *testing.T) {
		req := newServerHttpReq("/opds/v1.2/search")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>Search</ShortName>
  <Description>Search for Series</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <OutputEncoding>UTF-8</OutputEncoding>
  <Url template="/opds/v1.2/catalog?search={searchTerms}" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></Url>
</OpenSearchDescription>`, string(rec.Body.Bytes()))
	})
}

func TestServer_GetCatalog(t *testing.T) {
	emptyStore := mustOpenStoreMem(t)
	defer mustCloseStore(t, emptyStore)
	emptyRouter := router(emptyStore)

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/opds/v1.2/catalog", nil)
		rec := httptest.NewRecorder()
		emptyRouter.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("empty", func(t *testing.T) {
		req := newServerHttpReq("/opds/v1.2/catalog")
		rec := httptest.NewRecorder()
		emptyRouter.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<feed xmlns="http://www.w3.org/2005/Atom">
  <id>ctl</id>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/catalog" rel="self" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/search" rel="search" type="application/opensearchdescription+xml"></link>
  <title>Catalog</title>
  <updated>0001-01-01T00:00:00Z</updated>
  <author>
    <name>fiwippi</name>
    <uri>https://github.com/fiwippi</uri>
  </author>
</feed>`, string(rec.Body.Bytes()))
	})

	t.Run("not empty", func(t *testing.T) {
		r, s := newPopulatedRouter(t)
		defer mustCloseStore(t, s)

		req := newServerHttpReq("/opds/v1.2/catalog")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<feed xmlns="http://www.w3.org/2005/Atom">
  <id>ctl</id>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/catalog" rel="self" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/search" rel="search" type="application/opensearchdescription+xml"></link>
  <title>Catalog</title>
  <updated>2022-08-11T16:53:23+01:00</updated>
  <author>
    <name>fiwippi</name>
    <uri>https://github.com/fiwippi</uri>
  </author>
  <entry>
    <title>20th Century Boys</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI</id>
    <content></content>
    <link href="/opds/v1.2/series/PvHfuhL24GD6jo-PKLbPj_KvRikLn2WjCw_gOaXKRyI" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
  <entry>
    <title>Akira</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c</id>
    <content></content>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
  <entry>
    <title>Amano</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k</id>
    <content></content>
    <link href="/opds/v1.2/series/wNgocaIzfIjmFcxC-5I3S5pEpjRKjDY4nRxg9Ko-z7k" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
</feed>`, string(rec.Body.Bytes()))
	})

	t.Run("filtered", func(t *testing.T) {
		r, s := newPopulatedRouter(t)
		defer mustCloseStore(t, s)

		req := newServerHttpReq("/opds/v1.2/catalog?search=aki")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<feed xmlns="http://www.w3.org/2005/Atom">
  <id>ctl</id>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/catalog" rel="self" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/search" rel="search" type="application/opensearchdescription+xml"></link>
  <title>Catalog</title>
  <updated>2022-08-11T16:53:23+01:00</updated>
  <author>
    <name>fiwippi</name>
    <uri>https://github.com/fiwippi</uri>
  </author>
  <entry>
    <title>Akira</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c</id>
    <content></content>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
</feed>`, string(rec.Body.Bytes()))
	})
}

func TestServer_GetEntries(t *testing.T) {
	r, s := newPopulatedRouter(t)
	defer mustCloseStore(t, s)

	endpoint := "/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c"

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", endpoint, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid auth", func(t *testing.T) {
		req := newServerHttpReq(endpoint)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, `<feed xmlns="http://www.w3.org/2005/Atom">
  <id>rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c</id>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c" rel="self" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  <title>Akira</title>
  <updated>2022-08-11T16:53:23+01:00</updated>
  <author>
    <name>Katsuhiro Otomo</name>
    <uri></uri>
  </author>
  <entry>
    <title>Volume 01</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk</id>
    <content>zip - 26.3 KiB</content>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk/cover?thumbnail=true" rel="http://opds-spec.org/image/thumbnail" type="image/jpeg"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk/cover" rel="http://opds-spec.org/image" type="image/jpeg"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk/archive" rel="http://opds-spec.org/acquisition" type="application/zip"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/1f2Xo_TQk-nS-9I9QsRm3zVNawdW6HlOUYJsV22wENk/page/{pageNumber}" rel="http://vaemendis.net/opds-pse/stream" type="image/jpeg" xmlns:pse="http://vaemendis.net/opds-pse/ns" pse:count="15"></link>
  </entry>
  <entry>
    <title>Volume 02</title>
    <updated>2022-08-11T16:53:23+01:00</updated>
    <id>ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o</id>
    <content>zip - 18.3 KiB</content>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o/cover?thumbnail=true" rel="http://opds-spec.org/image/thumbnail" type="image/jpeg"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o/cover" rel="http://opds-spec.org/image" type="image/jpeg"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o/archive" rel="http://opds-spec.org/acquisition" type="application/zip"></link>
    <link href="/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o/page/{pageNumber}" rel="http://vaemendis.net/opds-pse/stream" type="image/jpeg" xmlns:pse="http://vaemendis.net/opds-pse/ns" pse:count="10"></link>
  </entry>
</feed>`, string(rec.Body.Bytes()))
	})
}

func TestServer_GetArchive(t *testing.T) {
	r, s := newPopulatedRouter(t)
	defer mustCloseStore(t, s)

	endpoint := "/opds/v1.2/series/rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c/entries/ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o/archive"

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", endpoint, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("valid auth", func(t *testing.T) {
		req := newServerHttpReq(endpoint)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		f, err := os.Open("tests/lib/Akira/Volume 02.zip")
		require.NoError(t, err)
		data, err := io.ReadAll(f)
		require.NoError(t, err)
		require.Equal(t, data, rec.Body.Bytes())
	})
}

func TestServer_GetCover(t *testing.T) {
	r, s := newPopulatedRouter(t)
	defer mustCloseStore(t, s)

	sid := "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c"
	eid := "ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o"
	endpoint := fmt.Sprintf("/opds/v1.2/series/%s/entries/%s/cover", sid, eid)

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", endpoint, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("original", func(t *testing.T) {
		req := newServerHttpReq(endpoint)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		buf, _, err := s.GetPage(sid, eid, 0)
		require.NoError(t, err)
		require.Equal(t, buf.Bytes(), rec.Body.Bytes())
	})

	t.Run("thumbnail", func(t *testing.T) {
		req := newServerHttpReq(endpoint + "?thumbnail=true")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusOK, rec.Code)

		buf, _, err := s.GetThumbnail(sid, eid)
		require.NoError(t, err)
		require.Equal(t, buf.Bytes(), rec.Body.Bytes())
	})
}

func TestServer_GetPage(t *testing.T) {
	r, s := newPopulatedRouter(t)
	defer mustCloseStore(t, s)

	sid := "rxogaPHmjap2Gwpwuo5K3EO7JYgxU21JCRuZBOvdc2c"
	eid := "ntnxQLqcSL5bQDAnFaRJKCqLMTjPtdqCEQZ1vipuw_o"
	endpoint := fmt.Sprintf("/opds/v1.2/series/%s/entries/%s/page/", sid, eid)

	t.Run("authorisation required", func(t *testing.T) {
		req := httptest.NewRequest("GET", endpoint+"0", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("all pages", func(t *testing.T) {
		entry, err := s.GetEntry(sid, eid)
		require.NoError(t, err)

		for i := range entry.Pages {
			req := newServerHttpReq(endpoint + strconv.Itoa(i))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			require.Equal(t, http.StatusOK, rec.Code)

			buf, _, err := s.GetPage(sid, eid, i)
			require.NoError(t, err)
			require.Equal(t, buf.Bytes(), rec.Body.Bytes())
		}
	})
}

func TestServer_ScanLibrary(t *testing.T) {
	conf := DefaultServerConfig()
	conf.ScanInterval = duration{time.Second}
	s := newTestServer(t, conf)
	require.NoError(t, s.Start())
	defer s.Stop()

	require.Eventually(t, func() bool {
		ctl, err := s.store.GetCatalog()
		require.NoError(t, err)
		return assert.ObjectsAreEqual([]Series{centurySeries, akiraSeries, amanoSeries}, ctl)
	}, 5*time.Second, time.Second)
}

// Utils

func newTestServer(t *testing.T, conf ServerConfig) *Server {
	conf.DataPath = InMemory
	conf.LibraryPath = "./tests/lib"
	s, err := NewServer(conf)
	require.NoError(t, err)
	return s
}

func newServerHttpReq(target string) *http.Request {
	req := httptest.NewRequest("GET", target, nil)
	req.SetBasicAuth(defaultUsername, defaultPassword)
	return req
}

func newPopulatedRouter(t *testing.T) (*chi.Mux, *Store) {
	s := mustOpenStoreMem(t)

	lib, err := ParseLibrary("tests/lib")
	require.NoError(t, err)
	require.NoError(t, s.PopulateCatalog(lib))

	return router(s), s
}
