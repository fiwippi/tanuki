package mangadex

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	c := New()
	l, err := c.Search(context.TODO(), "hori", 2)
	assert.Nil(t, err)

	chs, err := c.ViewChapters(context.TODO(), l[0])
	assert.Nil(t, err)
	fmt.Println(chs)
}
