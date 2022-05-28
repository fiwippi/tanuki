package manga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPages_Value(t *testing.T) {
	p := Pages{}
	data, err := p.Value()
	require.Nil(t, err)
	require.Equal(t, "[]", string(data.([]byte)))

	p = Pages{"HUH"}
	data, err = p.Value()
	require.Nil(t, err)
	require.Equal(t, "[\"HUH\"]", string(data.([]byte)))
}

func TestPages_Scan(t *testing.T) {
	p := Pages{}
	err := p.Scan([]byte("[]"))
	require.Nil(t, err)
	require.Equal(t, Pages{}, p)

	p = Pages{}
	err = p.Scan([]byte("[\"HUH\"]"))
	require.Nil(t, err)
	require.Equal(t, Pages{"HUH"}, p)
}
