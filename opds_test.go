package tanuki

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOPDS_Author(t *testing.T) {
	a := opdsAuthor{
		Name: "a",
		URI:  "b",
	}
	expected := `
<author>
  <name>a</name>
  <uri>b</uri>
</author>`

	b, err := xml.MarshalIndent(a, "", "  ")
	require.NoError(t, err)
	require.Equal(t, trimNewline(expected), string(b))
}

func TestOPDS_Time(t *testing.T) {
	t.Run("marshal XML", func(t *testing.T) {
		ti := opdsTime{time.Date(1999, 1, 1, 1, 1, 1, 1, time.UTC)}
		expected := `<opdsTime>1999-01-01T01:01:01.000000001Z</opdsTime>`

		b, err := xml.MarshalIndent(ti, "", "  ")
		require.NoError(t, err)
		require.Equal(t, expected, string(b))
	})
}

func TestOPDS_Search(t *testing.T) {
	t.Run("marshal XML", func(t *testing.T) {
		s := newOpdsSearch()
		expected := `
<OpenSearchDescription xmlns="http://a9.com/-/spec/opensearch/1.1/">
  <ShortName>Search</ShortName>
  <Description>Search for Series</Description>
  <InputEncoding>UTF-8</InputEncoding>
  <OutputEncoding>UTF-8</OutputEncoding>
  <Url template="/opds/v1.2/catalog?search={searchTerms}" type="application/atom+xml;profile=opds-catalog;kind=acquisition"></Url>
</OpenSearchDescription>`

		b, err := xml.MarshalIndent(s, "", "  ")
		require.NoError(t, err)
		require.Equal(t, trimNewline(expected), string(b))
	})
}

// Utils

func trimNewline(l string) string {
	l = strings.TrimPrefix(l, "\n")
	return strings.TrimSuffix(l, "\n")
}
