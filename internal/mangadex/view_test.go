package mangadex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViewManga(t *testing.T) {
	l, err := ViewManga(context.Background(), "4ec3e6b3-18bf-4964-9c87-1f287d3398f4")
	require.Nil(t, err)
	require.NotEqual(t, Listing{}, l)
}
