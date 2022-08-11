package feed

import (
	"encoding/xml"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTime_MarshalXML(t *testing.T) {
	ti := opdsTime{time.Date(1999, 1, 1, 1, 1, 1, 1, time.UTC)}
	expected := `<opdsTime>1999-01-01T01:01:01.000000001Z</opdsTime>`

	b, err := xml.MarshalIndent(ti, "", "  ")
	require.Nil(t, err)
	require.Equal(t, expected, string(b))
}
