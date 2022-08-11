package manga

import (
	"testing"

	"github.com/fiwippi/tanuki/internal/platform/image"
	"github.com/stretchr/testify/require"
)

func TestPages_Value(t *testing.T) {
	p := Pages{}
	require.Equal(t, 0, p.Total())
	data, err := p.Value()
	require.Nil(t, err)
	require.Equal(t, "[]", string(data.([]byte)))

	p = Pages{{Path: "HUH", Type: image.JPEG}}
	require.Equal(t, 1, p.Total())
	data, err = p.Value()
	require.Nil(t, err)
	require.Equal(t, `[{"path":"HUH","type":1}]`, string(data.([]byte)))
}

func TestPages_Scan(t *testing.T) {
	p := Pages{}
	err := p.Scan([]byte("[]"))
	require.Nil(t, err)
	require.Equal(t, Pages{}, p)
	require.Equal(t, 0, p.Total())

	p = Pages{}
	err = p.Scan([]byte(`[{"path":"HUH","type":1}]`))
	require.Nil(t, err)
	require.Equal(t, Pages{{Path: "HUH", Type: image.JPEG}}, p)
	require.Equal(t, 1, p.Total())
	require.Equal(t, image.JPEG, p[0].Type)
}
