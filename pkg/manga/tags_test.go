package manga

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTags_MarshalJSON(t *testing.T) {
	// Marshalling of empty set
	s := NewTags()
	data, err := s.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, "[]", string(data))
	val, err := s.Value()
	require.Nil(t, err)
	require.Equal(t, "[]", string(val.([]byte)))

	// Marshalling of set with elements
	s.Add("A")
	data, err = s.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, "[\"A\"]", string(data))
	val, err = s.Value()
	require.Nil(t, err)
	require.Equal(t, "[\"A\"]", string(val.([]byte)))
}

func TestTags_UnmarshalJSON(t *testing.T) {
	// Unmarshalling of empty set
	s := NewTags()
	err := s.UnmarshalJSON([]byte("[]"))
	require.Nil(t, err)
	require.True(t, s.Empty())
	require.Equal(t, 0, len(s.m))
	err = s.Scan([]byte("[]"))
	require.Nil(t, err)
	require.True(t, s.Empty())
	require.Equal(t, 0, len(s.m))

	// Unmarshalling of non-empty
	s = NewTags()
	err = s.UnmarshalJSON([]byte("[\"A\"]"))
	require.Nil(t, err)
	require.False(t, s.Empty())
	require.Equal(t, 1, len(s.m))
	require.True(t, s.Has("A"))
	err = s.Scan([]byte("[\"A\"]"))
	require.Nil(t, err)
	require.False(t, s.Empty())
	require.Equal(t, 1, len(s.m))
	require.True(t, s.Has("A"))
}
