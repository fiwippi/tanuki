package manga

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArchive_Exists(t *testing.T) {
	assert.False(t, new(Archive).Exists())
	a := Archive{Path: "../../tests/lib/Akira/Volume 01.zip"}
	assert.True(t, a.Exists())
}
