package feed

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/archive"
	"github.com/fiwippi/tanuki/pkg/manga"
)

func TestSeries(t *testing.T) {
	t1 := time.Date(1, 1, 1, 1, 1, 1, 0, time.UTC)
	t2 := time.Date(2, 1, 1, 1, 1, 1, 0, time.UTC)

	s := NewSeriesFeed(opdsRoot, "a", "b")
	s.SetAuthor("a", "b")
	s.SetUpdated(t1)
	s.AddEntry("c", "d", "e", "f", "g", 5, t2, &manga.Archive{
		Title: "h",
		Path:  "i",
		Type:  archive.Zip,
	})

	expected := `
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>a</id>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/series/a" rel="self" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  <title>b</title>
  <updated>0001-01-01T01:01:01Z</updated>
  <author>
    <name>a</name>
    <uri>b</uri>
  </author>
  <entry>
    <title>d</title>
    <updated>0002-01-01T01:01:01Z</updated>
    <id>c</id>
    <content>zip - 0.0 MiB</content>
    <link href="/opds/v1.2/series/a/entries/c/cover?thumbnail=true" rel="http://opds-spec.org/image/thumbnail" type="e"></link>
    <link href="/opds/v1.2/series/a/entries/c/cover" rel="http://opds-spec.org/image" type="f"></link>
    <link href="/opds/v1.2/series/a/entries/c/archive" rel="http://opds-spec.org/acquisition" type="application/zip"></link>
    <link href="/opds/v1.2/series/a/entries/c/page/{pageNumber}?zero_based=true" rel="http://vaemendis.net/opds-pse/stream" type="g" xmlns:pse="http://vaemendis.net/opds-pse/ns" pse:count="5"></link>
  </entry>
</feed>`

	b, err := xml.MarshalIndent(s, "", "  ")
	require.Nil(t, err)
	require.Equal(t, trimNewline(expected), string(b))
}
