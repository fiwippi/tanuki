package progress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const emptySP = "{}"
const nonEmptySP = "{\"a\":{\"eid\":\"a\",\"current\":1,\"total\":5}}"

func TestSeriesProgress_MarshalJSON(t *testing.T) {
	// Marshalling of empty series progress
	sp := NewSeriesProgress()
	data, err := sp.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, emptySP, string(data))

	// Marshalling of series progress with entries
	e := Entry{
		EID:     "a",
		Current: 1,
		Total:   5,
	}
	sp.Add(e.EID, e)
	data, err = sp.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, nonEmptySP, string(data))
}

func TestSeriesProgress_UnmarshalJSON(t *testing.T) {
	// Unmarshalling of empty series progress
	sp := NewSeriesProgress()
	err := sp.UnmarshalJSON([]byte(emptySP))
	require.Nil(t, err)
	require.NotNil(t, sp.m)
	require.Equal(t, 0, len(sp.m))

	// Unmarshalling of series progress with entries
	sp = NewSeriesProgress()
	err = sp.UnmarshalJSON([]byte(nonEmptySP))
	require.Nil(t, err)
	require.NotNil(t, sp.m)
	require.Equal(t, 1, len(sp.m))
	ep, found := sp.m["a"]
	require.True(t, found)
	require.NotEqual(t, Entry{}, ep)
	require.Equal(t, "a", ep.EID)
	require.Equal(t, 1, ep.Current)
	require.Equal(t, 5, ep.Total)
}
