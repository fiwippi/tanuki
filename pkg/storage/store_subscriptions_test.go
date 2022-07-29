package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
)

func setGetSubscription(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	seriesA := parsedData[0].s
	seriesB := parsedData[1].s
	require.Nil(t, s.AddSeries(seriesA, parsedData[0].e))
	require.Nil(t, s.AddSeries(seriesB, parsedData[1].e))

	// Subscription with no time
	require.Nil(t, s.SetSubscription(seriesA.SID, "a", "a", true))
	sub, err := s.GetSubscription(seriesA.SID)
	require.Nil(t, err)
	require.Equal(t, seriesA.SID, sub.SID)
	require.Equal(t, "a", sub.Title)
	require.Equal(t, dbutil.NullString("a"), sub.MdexUUID)
	require.True(t, sub.MdexLastPublishedAt.Equal(dbutil.Time{}))

	// Subscription with set time
	tiOlder := dbutil.Time(time.Now())
	require.Nil(t, s.SetSubscriptionWithTime(seriesB.SID, "b", "b", tiOlder, true))
	sub, err = s.GetSubscription(seriesB.SID)
	require.Nil(t, err)
	require.Equal(t, seriesB.SID, sub.SID)
	require.Equal(t, "b", sub.Title)
	require.Equal(t, dbutil.NullString("b"), sub.MdexUUID)
	require.True(t, sub.MdexLastPublishedAt.Equal(tiOlder))

	// Newer subscription should overwrite the older one
	tiNewer := dbutil.Time(tiOlder.Time().Add(1 * time.Hour))
	require.Nil(t, s.SetSubscriptionWithTime(seriesB.SID, "b", "b", tiNewer, true))
	sub, err = s.GetSubscription(seriesB.SID)
	require.Nil(t, err)
	require.Equal(t, seriesB.SID, sub.SID)
	require.Equal(t, "b", sub.Title)
	require.Equal(t, dbutil.NullString("b"), sub.MdexUUID)
	require.True(t, sub.MdexLastPublishedAt.Equal(tiNewer))

	// Setting the older subscription should still show the newer subscription
	require.Nil(t, s.SetSubscriptionWithTime(seriesB.SID, "b", "b", tiOlder, true))
	sub, err = s.GetSubscription(seriesB.SID)
	require.Nil(t, err)
	require.Equal(t, seriesB.SID, sub.SID)
	require.Equal(t, "b", sub.Title)
	require.Equal(t, dbutil.NullString("b"), sub.MdexUUID)
	require.True(t, sub.MdexLastPublishedAt.Equal(tiNewer))

	// Setting the older subscription without ensure newest should now show the older subscription
	require.Nil(t, s.SetSubscriptionWithTime(seriesB.SID, "b", "b", tiOlder, false))
	sub, err = s.GetSubscription(seriesB.SID)
	require.Nil(t, err)
	require.Equal(t, seriesB.SID, sub.SID)
	require.Equal(t, "b", sub.Title)
	require.Equal(t, dbutil.NullString("b"), sub.MdexUUID)
	require.True(t, sub.MdexLastPublishedAt.Equal(tiOlder))
}

func getAllDeleteSubscriptions(t *testing.T) {
	s := mustOpenStoreMem(t)
	defer mustCloseStore(t, s)

	seriesA := parsedData[0].s
	seriesB := parsedData[1].s
	require.Nil(t, s.AddSeries(seriesA, parsedData[0].e))
	require.Nil(t, s.AddSeries(seriesB, parsedData[1].e))

	// No subscriptions should return no error
	sub, err := s.GetAllSubscriptions()
	assert.Nil(t, err)
	assert.Zero(t, len(sub))

	// 2 subscriptions should be returned
	require.Nil(t, s.SetSubscription(seriesA.SID, "a", "a", true))
	require.Nil(t, s.SetSubscription(seriesB.SID, "b", "b", true))
	sub, err = s.GetAllSubscriptions()
	assert.Nil(t, err)
	assert.Len(t, sub, 2)

	// Delete subscriptions
	require.Nil(t, s.DeleteSubscription(seriesA.SID))
	require.Nil(t, s.DeleteSubscription(seriesB.SID))

	// Should have no subscriptions now
	sub, err = s.GetAllSubscriptions()
	assert.Nil(t, err)
	assert.Zero(t, len(sub))
}

func TestStore_SetSubscription(t *testing.T) {
	setGetSubscription(t)
}

func TestStore_SetSubscriptionWithTime(t *testing.T) {
	setGetSubscription(t)
}

func TestStore_GetSubscription(t *testing.T) {
	setGetSubscription(t)
}

func TestStore_GetAllSubscriptions(t *testing.T) {
	getAllDeleteSubscriptions(t)
}

func TestStore_DeleteSubscription(t *testing.T) {
	getAllDeleteSubscriptions(t)
}
