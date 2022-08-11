package storage

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/pkg/human"
)

func TestStore_SetEntryProgressAmount(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entry := parsedData[0].e[0]
	require.Nil(t, s.AddSeries(series, parsedData[0].e))
	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	require.Nil(t, s.SetEntryProgressAmount(series.SID, entry.EID, u.UID, 5))
	p, err := s.GetEntryProgress(series.SID, entry.EID, u.UID)
	require.Nil(t, err)
	require.Equal(t, entry.EID, p.EID)
	require.Equal(t, 5, p.Current)
	require.Equal(t, p.Total, len(entry.Pages))

	mustCloseStore(t, s)
}

func TestStore_SetEntryProgressRead(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entry := parsedData[0].e[0]
	require.Nil(t, s.AddSeries(series, parsedData[0].e))
	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	require.Nil(t, s.SetEntryProgressRead(series.SID, entry.EID, u.UID))
	p, err := s.GetEntryProgress(series.SID, entry.EID, u.UID)
	require.Nil(t, err)
	require.Equal(t, entry.EID, p.EID)
	require.Equal(t, len(entry.Pages), p.Current)
	require.Equal(t, len(entry.Pages), p.Total)

	mustCloseStore(t, s)
}

func TestStore_SetEntryProgressUnread(t *testing.T) {
	s := mustOpenStoreMem(t)

	series := parsedData[0].s
	entry := parsedData[0].e[0]
	require.Nil(t, s.AddSeries(series, parsedData[0].e))
	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	require.Nil(t, s.SetEntryProgressUnread(series.SID, entry.EID, u.UID))
	p, err := s.GetEntryProgress(series.SID, entry.EID, u.UID)
	require.Nil(t, err)
	require.Equal(t, entry.EID, p.EID)
	require.Equal(t, 0, p.Current)
	require.Equal(t, len(entry.Pages), p.Total)

	mustCloseStore(t, s)
}

func TestStore_SetSeriesProgressRead(t *testing.T) {
	s := mustOpenStoreMem(t)

	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
		require.Nil(t, s.SetSeriesProgressRead(d.s.SID, u.UID))

		sp, err := s.GetSeriesProgress(d.s.SID, u.UID)
		require.Nil(t, err)
		require.NotNil(t, sp)

		for _, e := range d.e {
			// Check each entry can be accessed individually from the store
			p1, err := s.GetEntryProgress(d.s.SID, e.EID, u.UID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p1.EID)
			require.Equal(t, len(e.Pages), p1.Current)
			require.Equal(t, len(e.Pages), p1.Total)

			// Check the entry progress can be retrieved from the series progress
			p2, err := sp.Get(e.EID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p2.EID)
			require.Equal(t, len(e.Pages), p2.Current)
			require.Equal(t, len(e.Pages), p2.Total)

			require.Equal(t, p1, p2)
		}
	}

	mustCloseStore(t, s)
}

func TestStore_SetSeriesProgressUnread(t *testing.T) {
	s := mustOpenStoreMem(t)

	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	for _, d := range parsedData {
		require.Nil(t, s.AddSeries(d.s, d.e))
		require.Nil(t, s.SetSeriesProgressUnread(d.s.SID, u.UID))

		sp, err := s.GetSeriesProgress(d.s.SID, u.UID)
		require.Nil(t, err)
		require.NotNil(t, sp)

		for _, e := range d.e {
			// Check each entry can be accessed individually from the store
			p1, err := s.GetEntryProgress(d.s.SID, e.EID, u.UID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p1.EID)
			require.Equal(t, 0, p1.Current)
			require.Equal(t, len(e.Pages), p1.Total)

			// Check the entry progress can be retrieved from the series progress
			p2, err := sp.Get(e.EID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p2.EID)
			require.Equal(t, 0, p2.Current)
			require.Equal(t, len(e.Pages), p2.Total)

			require.Equal(t, p1, p2)
		}
	}

	mustCloseStore(t, s)
}

func TestStore_GetEntryProgress(t *testing.T) {
	s := mustOpenStoreMem(t)

	require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	t.Run("ProgressDoesNotExist", func(t *testing.T) {
		// Getting progress for an entry which has no
		// progress attached to it should return empty
		// progress for it
		entry := parsedData[0].e[0]
		p, err := s.GetEntryProgress(entry.SID, entry.EID, u.UID)
		require.Nil(t, err)
		require.NotEqual(t, human.EntryProgress{}, p)
		require.Equal(t, entry.EID, p.EID)
		require.Equal(t, 0, p.Current)
		require.Equal(t, len(entry.Pages), p.Total)
	})

	t.Run("ProgressExists", func(t *testing.T) {
		entry := parsedData[0].e[1]
		require.Nil(t, s.SetEntryProgressAmount(entry.SID, entry.EID, u.UID, 5))
		p, err := s.GetEntryProgress(entry.SID, entry.EID, u.UID)
		require.Nil(t, err)
		require.NotEqual(t, human.EntryProgress{}, p)
		require.Equal(t, entry.EID, p.EID)
		require.Equal(t, 5, p.Current)
		require.Equal(t, len(entry.Pages), p.Total)
	})

	mustCloseStore(t, s)
}

func TestStore_GetSeriesProgress(t *testing.T) {
	s := mustOpenStoreMem(t)

	require.Nil(t, s.AddSeries(parsedData[0].s, parsedData[0].e))
	require.Nil(t, s.AddSeries(parsedData[1].s, parsedData[1].e))
	u := human.NewUser("a", "a", human.Standard)
	require.Nil(t, s.AddUser(u, true))

	t.Run("ProgressDoesNotExist", func(t *testing.T) {
		// Should return an empty series progress
		series := parsedData[0].s
		sp, err := s.GetSeriesProgress(series.SID, u.UID)
		require.Nil(t, err)
		require.Equal(t, *human.NewSeriesProgress(), sp)
	})

	t.Run("ProgressExists", func(t *testing.T) {
		series := parsedData[1].s
		entries := parsedData[1].e
		require.Nil(t, s.SetSeriesProgressUnread(series.SID, u.UID))
		sp, err := s.GetSeriesProgress(series.SID, u.UID)
		require.Nil(t, err)
		require.NotNil(t, sp)

		for _, e := range entries {
			// Check each entry can be accessed individually from the store
			p1, err := s.GetEntryProgress(series.SID, e.EID, u.UID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p1.EID)
			require.Equal(t, 0, p1.Current)
			require.Equal(t, len(e.Pages), p1.Total)

			// Check the entry progress can be retrieved from the series progress
			p2, err := sp.Get(e.EID)
			require.Nil(t, err)
			require.Equal(t, e.EID, p2.EID)
			require.Equal(t, 0, p2.Current)
			require.Equal(t, len(e.Pages), p2.Total)

			require.Equal(t, p1, p2)
		}
	})

	mustCloseStore(t, s)
}

func TestStore_ProgressDeletedOnCascade(t *testing.T) {
	t.Run("DeletedOnUserDelete", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		u := human.NewUser("a", "a", human.Standard)
		require.Nil(t, s.AddUser(u, true))

		for _, d := range parsedData {
			require.Nil(t, s.AddSeries(d.s, d.e))
			require.Nil(t, s.SetSeriesProgressUnread(d.s.SID, u.UID))
		}

		var exists bool

		require.Nil(t, s.pool.Get(&exists, `SELECT COUNT(*) > 0 FROM progress`))
		require.True(t, exists)

		require.Nil(t, s.DeleteUser(u.UID))
		has, err := s.HasUser(u.UID)
		require.Nil(t, err)
		require.False(t, has)

		require.Nil(t, s.pool.Get(&exists, `SELECT COUNT(*) > 0 FROM progress`))
		require.False(t, exists)
	})

	t.Run("DeletedOnEntryDelete", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		u := human.NewUser("a", "a", human.Standard)
		require.Nil(t, s.AddUser(u, true))

		for _, d := range parsedData {
			require.Nil(t, s.AddSeries(d.s, d.e))
			require.Nil(t, s.SetSeriesProgressUnread(d.s.SID, u.UID))

			for _, e := range d.e {
				var exists bool
				stmt := `SELECT COUNT(*) > 0 FROM progress WHERE sid = ? AND eid = ?`

				require.Nil(t, s.pool.Get(&exists, stmt, d.s.SID, e.EID))
				require.True(t, exists)

				fn := func(tx *sqlx.Tx) error {
					return s.deleteEntry(tx, d.s.SID, e.EID)
				}
				require.Nil(t, s.tx(fn))

				require.Nil(t, s.pool.Get(&exists, stmt, d.s.SID, e.EID))
				require.False(t, exists)
			}
		}
	})

	t.Run("DeletedOnSeriesDelete", func(t *testing.T) {
		s := mustOpenStoreMem(t)
		defer mustCloseStore(t, s)

		u := human.NewUser("a", "a", human.Standard)
		require.Nil(t, s.AddUser(u, true))

		for _, d := range parsedData {
			require.Nil(t, s.AddSeries(d.s, d.e))
			require.Nil(t, s.SetSeriesProgressUnread(d.s.SID, u.UID))

			var exists bool
			stmt := `SELECT COUNT(*) > 0 FROM progress WHERE sid = ? AND eid = ?`

			for _, e := range d.e {
				require.Nil(t, s.pool.Get(&exists, stmt, d.s.SID, e.EID))
				require.True(t, exists)
			}

			fn := func(tx *sqlx.Tx) error {
				return s.deleteSeries(tx, d.s.SID)
			}
			require.Nil(t, s.tx(fn))

			for _, e := range d.e {
				require.Nil(t, s.pool.Get(&exists, stmt, d.s.SID, e.EID))
				require.False(t, exists)
			}
		}
	})
}
