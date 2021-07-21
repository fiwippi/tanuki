package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

// Editing

func (db *DB) CreateUser(u *core.User) error {
	return db.SaveUser(u, false)
}

func (db *DB) SaveUser(u *core.User, overwrite bool) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)
		return root.AddUser(u, overwrite)
	})
}

func (db *DB) DeleteUser(uid string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return db.usersBucket(tx).DeleteUser(uid)
	})
}

func (db *DB) ChangeUsername(uid, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)
		return root.RenameUser(uid, auth.HashSHA1(name), name)
	})
}

func (db *DB) ChangePassword(uid, password string) error {
	return db.Update(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}
		return user.ChangePassword(password, true)
	})
}

func (db *DB) ChangeUserType(uid string, uType core.UserType) error {
	admins := db.AdminCount()
	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}

		// Ensure 1 admin will exist by the end of the transaction
		if user.Type() == core.AdminUser && uType != core.AdminUser {
			admins -= 1
		}
		if admins == 0 {
			return ErrNotEnoughAdmins
		}

		// Set the password
		return user.ChangeType(uType)
	})
}

func (db *DB) ChangeProgressTracker(uid string, t *core.ProgressTracker) error {
	return db.Update(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}
		return user.ChangeProgressTracker(t)
	})
}

// TODO move progress data into the series bucket
func (db *DB) SetSeriesProgressAllRead(uid, sid string) error {
	u, err := db.GetUser(uid)
	if err != nil {
		return err
	}

	// Check series exists and ensure the series entry count
	// is the same as the read progress entry count
	entries, err := db.GetSeriesEntries(sid)
	if err != nil {
		return err
	}
	num, err := u.ProgressTracker.SeriesEntriesNum(sid)
	if err != nil {
		return err
	}
	if len(entries) != num {
		return ErrProgressCount
	}

	// Change the series progress
	err = u.ProgressTracker.SetSeriesAllRead(sid)
	if err != nil {
		return err
	}

	// Save the new progress
	return db.ChangeProgressTracker(uid, u.ProgressTracker)
}

func (db *DB) SetSeriesProgressAllUnread(uid, sid string) error {
	u, err := db.GetUser(uid)
	if err != nil {
		return err
	}

	// Check series exists and ensure the series entry count
	// is the same as the read progress entry count
	entries, err := db.GetSeriesEntries(sid)
	if err != nil {
		return err
	}
	num, err := u.ProgressTracker.SeriesEntriesNum(sid)
	if err != nil {
		return err
	}
	if len(entries) != num {
		return ErrProgressCount
	}

	// Change the series progress
	err = u.ProgressTracker.SetSeriesAllUnread(sid)
	if err != nil {
		return err
	}

	// Save the new progress
	return db.ChangeProgressTracker(uid, u.ProgressTracker)
}

func (db *DB) SetEntryProgressRead(uid, sid, eid string) error {
	u, err := db.GetUser(uid)
	if err != nil {
		return err
	}

	// Check series exists and the entry exists
	entry, err := db.GetEntry(sid, eid)
	if err != nil {
		return err
	}

	// Ensure progress tracker is tracking this entry
	if !u.ProgressTracker.HasEntry(sid, eid) {
		u.ProgressTracker.AddEntry(sid, eid, entry.Pages)
	}

	// Set progress
	u.ProgressTracker.ProgressEntry(sid, eid).SetRead()

	// Save the new progress
	return db.ChangeProgressTracker(uid, u.ProgressTracker)
}

func (db *DB) SetEntryProgressUnread(uid, sid, eid string) error {
	u, err := db.GetUser(uid)
	if err != nil {
		return err
	}

	// Check series exists and the entry exists
	entry, err := db.GetEntry(sid, eid)
	if err != nil {
		return err
	}

	// Ensure progress tracker is tracking this entry
	if !u.ProgressTracker.HasEntry(sid, eid) {
		u.ProgressTracker.AddEntry(sid, eid, entry.Pages)
	}

	// Set progress
	u.ProgressTracker.ProgressEntry(sid, eid).SetUnread()

	// Save the new progress
	return db.ChangeProgressTracker(uid, u.ProgressTracker)
}

func (db *DB) SetSeriesEntryProgressNum(uid, sid, eid string, num int) error {
	u, err := db.GetUser(uid)
	if err != nil {
		return err
	}

	// Check series exists and the entry exists
	entry, err := db.GetEntry(sid, eid)
	if err != nil {
		return err
	}

	// Ensure progress tracker is tracking this entry
	if !u.ProgressTracker.HasEntry(sid, eid) {
		u.ProgressTracker.AddEntry(sid, eid, entry.Pages)
	}

	// Set progress
	u.ProgressTracker.ProgressEntry(sid, eid).Set(num)

	// Save the new progress
	return db.ChangeProgressTracker(uid, u.ProgressTracker)
}

// Viewing

func (db *DB) GetUser(uid string) (*core.User, error) {
	var e *core.User
	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}
		e = user.Struct()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (db *DB) GetUsername(uid string) (string, error) {
	var e string

	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}
		e = user.Name()

		return nil
	})
	if err != nil {
		return "", err
	}

	return e, nil
}

func (db *DB) GetUsers(safe bool) ([]core.User, error) {
	users := make([]core.User, 0)
	err := db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		return root.ForEachUser(func(ub *UserBucket) error {
			u := ub.Struct()
			if safe {
				u.Pass = ""
			}
			users = append(users, *u)
			return nil
		})
	})

	return users, err
}

func (db *DB) HasUser(uid string) bool {
	exists := false
	db.View(func(tx *bolt.Tx) error {
		exists = db.usersBucket(tx).HasUser(uid)
		return nil
	})

	return exists
}

func (db *DB) HasUsers() bool {
	exists := false
	db.View(func(tx *bolt.Tx) error {
		exists = db.usersBucket(tx).HasUsers()
		return nil
	})

	return exists
}

// Processing

func (db *DB) IsUserAdmin(uid string) (bool, error) {
	var isAdmin bool

	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}
		isAdmin = user.IsAdmin()

		return nil
	})
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

func (db *DB) AdminCount() int {
	var count int
	db.View(func(tx *bolt.Tx) error {
		count = db.usersBucket(tx).AdminCount()
		return nil
	})
	return count
}

func (db *DB) ValidateLogin(name, pass string) (bool, error) {
	var valid bool

	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(auth.HashSHA1(name))
		if err != nil {
			return err
		}
		valid = user.ValidPassword(pass)

		return nil
	})
	if err != nil {
		return false, err
	}

	return valid, nil
}

func (db *DB) EnsureValidSeriesProgress(sid, eid string, num int) {
	db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		return root.ForEachUser(func(ub *UserBucket) error {
			tracker := ub.ProgressTracker()
			if tracker != nil {
				e := tracker.ProgressEntry(sid, eid)
				if e != nil {
					e.Total = num
					ub.ChangeProgressTracker(tracker)
				}
			}

			return nil
		})
	})
}
