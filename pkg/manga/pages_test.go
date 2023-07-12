package manga

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/image"
)

func TestPages_Value(t *testing.T) {
	p := Pages{}
	require.Equal(t, 0, len(p))
	data, err := p.Value()
	require.Nil(t, err)
	require.Equal(t, "[]", string(data.([]byte)))

	p = Pages{{Path: "HUH", Type: image.JPEG}}
	require.Equal(t, 1, len(p))
	data, err = p.Value()
	require.Nil(t, err)
	require.Equal(t, `[{"path":"HUH","type":1}]`, string(data.([]byte)))
}

func TestPages_Scan(t *testing.T) {
	p := Pages{}
	err := p.Scan([]byte("[]"))
	require.Nil(t, err)
	require.Equal(t, Pages{}, p)
	require.Equal(t, 0, len(p))

	p = Pages{}
	err = p.Scan([]byte(`[{"path":"HUH","type":1}]`))
	require.Nil(t, err)
	require.Equal(t, Pages{{Path: "HUH", Type: image.JPEG}}, p)
	require.Equal(t, 1, len(p))
	require.Equal(t, image.JPEG, p[0].Type)
}
