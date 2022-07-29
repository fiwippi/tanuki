package mangadex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateManga(t *testing.T) {
	valid, err := ValidateManga(context.Background(), "a25e46ec-30f7-4db6-89df-cacbc1d9a900")
	assert.Nil(t, err)
	assert.True(t, valid)

	valid, err = ValidateManga(context.Background(), "xxx")
	assert.NotNil(t, err)
	assert.False(t, valid)
}
