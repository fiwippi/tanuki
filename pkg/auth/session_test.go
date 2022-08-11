package auth

import (
	"testing"
	"time"

	"github.com/shaj13/libcache"
	"github.com/stretchr/testify/require"
)

func TestKeysDeleteOnExpiry(t *testing.T) {
	cache := libcache.LRU.New(10)
	cache.SetTTL(100 * time.Millisecond)
	send := make(chan libcache.Event, 10)
	cache.Notify(send, libcache.Read, libcache.Write, libcache.Remove)

	cache.Store(1, 0)
	_, exists := cache.Load(1)
	require.True(t, exists)
	time.Sleep(200 * time.Millisecond)
	_, exists = cache.Load(1)
	require.False(t, exists)
}
