package config

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/fse"
)

func TestPaths_EnsureExist(t *testing.T) {
	p := Paths{
		DB:      "./a/a.txt",
		Log:     "./b/b.txt",
		Library: "./c",
	}

	var wg sync.WaitGroup
	wg.Add(2)

	t.Run("LogDirNotCreated", func(t *testing.T) {
		defer func() {
			defer wg.Done()
			require.Nil(t, os.Remove("./a"))
			require.Nil(t, os.Remove("./c"))
		}()

		require.Nil(t, p.EnsureExist(false))
		require.True(t, fse.Exists("./a"))
		require.False(t, fse.Exists("./b"))
		require.True(t, fse.Exists("./c"))
	})

	t.Run("LogDirCreated", func(t *testing.T) {
		defer func() {
			defer wg.Done()
			require.Nil(t, os.Remove("./a"))
			require.Nil(t, os.Remove("./b"))
			require.Nil(t, os.Remove("./c"))
		}()

		require.Nil(t, p.EnsureExist(true))
		require.True(t, fse.Exists("./a"))
		require.True(t, fse.Exists("./b"))
		require.True(t, fse.Exists("./c"))
	})

	wg.Wait()
}
