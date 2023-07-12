package feed

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const opdsRoot = "/opds/v1.2"

func TestCatalog(t *testing.T) {
	ctl := NewCatalogFeed(opdsRoot)
	ctl.SetAuthor("a", "b")
	ctl.SetUpdated(time.Date(1, 1, 1, 1, 1, 1, 0, time.UTC))
	ctl.AddSeries("c", "c", time.Date(2, 1, 1, 1, 1, 1, 0, time.UTC))
	ctl.AddSeries("d", "d", time.Date(3, 1, 1, 1, 1, 1, 0, time.UTC))
	ctl.AddSeries("e", "e", time.Date(4, 1, 1, 1, 1, 1, 0, time.UTC))

	expected := `
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>root</id>
  <link href="/opds/v1.2/catalog" rel="self" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <link href="/opds/v1.2/catalog" rel="start" type="application/atom+xml;profile=opds-catalog;kind=navigation"></link>
  <title>Catalog</title>
  <updated>0001-01-01T01:01:01Z</updated>
  <author>
    <name>a</name>
    <uri>b</uri>
  </author>
  <entry>
    <title>c</title>
    <updated>0002-01-01T01:01:01Z</updated>
    <id>c</id>
    <content></content>
    <link href="/opds/v1.2/series/c" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
  <entry>
    <title>d</title>
    <updated>0003-01-01T01:01:01Z</updated>
    <id>d</id>
    <content></content>
    <link href="/opds/v1.2/series/d" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
  <entry>
    <title>e</title>
    <updated>0004-01-01T01:01:01Z</updated>
    <id>e</id>
    <content></content>
    <link href="/opds/v1.2/series/e" rel="subsection" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></link>
  </entry>
</feed>`

	b, err := xml.MarshalIndent(ctl, "", "  ")
	require.Nil(t, err)
	require.Equal(t, trimNewline(expected), string(b))
}
