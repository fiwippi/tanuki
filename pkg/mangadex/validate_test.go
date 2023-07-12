package mangadex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateManga(t *testing.T) {
	valid, err := ValidateManga(context.Background(), "a25e46ec-30f7-4db6-89df-cacbc1d9a900")
	require.Nil(t, err)
	require.True(t, valid)

	valid, err = ValidateManga(context.Background(), "xxx")
	require.NotNil(t, err)
	require.False(t, valid)
}
