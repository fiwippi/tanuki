package storage

import (
	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/platform/dbutil"
	"github.com/fiwippi/tanuki/pkg/manga"
)

func (s *Store) setSubscriptionWithTime(tx *sqlx.Tx, sid, title string, uuid dbutil.NullString, t dbutil.Time, ensureNewest bool) error {
	sb, err := s.getSubscription(tx, sid)

	// Always set if not ensure newest
	a := !ensureNewest
	// Set if error occurred meaning there is no previous subscription
	b := ensureNewest && err != nil
	// If there is a previous subscription then set if the time is newer
	c := ensureNewest && err == nil && t.After(sb.MdexLastPublishedAt)

	if a || b || c {
		stmt := `
			REPLACE INTO subscriptions 
				(sid, title, mangadex_uuid, mangadex_last_published_at) 
			Values 
				(?, ?, ?, ?)`
		if _, err := tx.Exec(stmt, sid, title, uuid, t); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SetSubscription(sid, title string, uuid dbutil.NullString, ensureNewest bool) error {
	return s.SetSubscriptionWithTime(sid, title, uuid, dbutil.Time{}, ensureNewest)
}

func (s *Store) SetSubscriptionWithTime(sid, title string, uuid dbutil.NullString, t dbutil.Time, ensureNewest bool) error {
	fn := func(tx *sqlx.Tx) error {
		return s.setSubscriptionWithTime(tx, sid, title, uuid, t, ensureNewest)
	}

	if err := s.tx(fn); err != nil {
		return err
	}
	return nil
}

func (s *Store) getSubscription(tx *sqlx.Tx, sid string) (manga.Subscription, error) {
	var sb manga.Subscription
	if err := tx.Get(&sb, `SELECT * FROM subscriptions WHERE sid = ?`, sid); err != nil {
		return manga.Subscription{}, err
	}
	return sb, nil
}

func (s *Store) GetSubscription(sid string) (manga.Subscription, error) {
	var sb manga.Subscription
	fn := func(tx *sqlx.Tx) error {
		var err error
		sb, err = s.getSubscription(tx, sid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return manga.Subscription{}, err
	}
	return sb, nil
}

func (s *Store) getAllSubscriptions(tx *sqlx.Tx) ([]manga.Subscription, error) {
	var sb []manga.Subscription
	if err := tx.Select(&sb, `SELECT * FROM subscriptions`); err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *Store) GetAllSubscriptions() ([]manga.Subscription, error) {
	var sb []manga.Subscription
	fn := func(tx *sqlx.Tx) error {
		var err error
		sb, err = s.getAllSubscriptions(tx)
		return err
	}

	if err := s.tx(fn); err != nil {
		return nil, err
	}
	return sb, nil
}

func (s *Store) deleteSubscription(tx *sqlx.Tx, sid string) error {
	_, err := tx.Exec(`DELETE FROM subscriptions WHERE sid = ?`, sid)
	return err
}

func (s *Store) DeleteSubscription(sid string) error {
	fn := func(tx *sqlx.Tx) error {
		return s.deleteSubscription(tx, sid)
	}

	if err := s.tx(fn); err != nil {
		return err
	}
	return nil
}
