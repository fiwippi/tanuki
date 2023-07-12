package feed

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func trimNewline(l string) string {
	l = strings.TrimPrefix(l, "\n")
	return strings.TrimSuffix(l, "\n")
}

func TestAuthor(t *testing.T) {
	a := author{
		Name: "a",
		URI:  "b",
	}
	expected := `
<author>
  <name>a</name>
  <uri>b</uri>
</author>`

	b, err := xml.MarshalIndent(a, "", "  ")
	require.Nil(t, err)
	require.Equal(t, trimNewline(expected), string(b))
}
