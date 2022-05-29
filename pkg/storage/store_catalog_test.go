package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore_GenerateThumbnails(t *testing.T) {
	s := mustOpenStoreMem(t)

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
	}
	//require.Nil(t, s.GenerateThumbnails(true))

	mustCloseStore(t, s)
}
