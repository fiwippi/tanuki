package manga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPages_Value(t *testing.T) {
	p := Pages{}
	require.Equal(t, 0, p.Total())
	data, err := p.Value()
	require.Nil(t, err)
	require.Equal(t, "[]", string(data.([]byte)))

	p = Pages{"HUH"}
	require.Equal(t, 1, p.Total())
	data, err = p.Value()
	require.Nil(t, err)
	require.Equal(t, "[\"HUH\"]", string(data.([]byte)))
}

func TestPages_Scan(t *testing.T) {
	p := Pages{}
	err := p.Scan([]byte("[]"))
	require.Nil(t, err)
	require.Equal(t, Pages{}, p)
	require.Equal(t, 0, p.Total())

	p = Pages{}
	err = p.Scan([]byte("[\"HUH\"]"))
	require.Nil(t, err)
	require.Equal(t, Pages{"HUH"}, p)
	require.Equal(t, 1, p.Total())
}
