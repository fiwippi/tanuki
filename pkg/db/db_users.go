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

func (db *DB) DeleteUserHashed(hash string) error {
	return db.deleteUser([]byte(hash))
}

func (db *DB) DeleteUserUnhashed(name string) error {
	return db.deleteUser([]byte(auth.HashSHA1(name)))
}

func (db *DB) deleteUser(hash []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)
		return root.DeleteUser(hash)
	})
}

func (db *DB) ChangeUserNameHashed(oldHash, newHash string, newUsername string) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)
		return root.RenameUser(oldHash, newHash, newUsername)
	})
}

func (db *DB) ChangeUserNameUnhashed(oldName, newName string) error {
	return db.ChangeUserNameHashed(auth.HashSHA1(oldName), auth.HashSHA1(newName), newName)
}

func (db *DB) ChangeUserPasswordHashed(hash, unhashedPassword string) error {
	return db.Update(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
		if err != nil {
			return err
		}
		return user.ChangePassword(unhashedPassword)
	})
}

func (db *DB) ChangeUserPasswordUnhashed(name, password string) error {
	return db.ChangeUserPasswordHashed(auth.HashSHA1(name), password)
}

func (db *DB) ChangeUserTypeHashed(hash string, uType core.UserType) error {
	admins, err := db.AdminCount()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
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

func (db *DB) ChangeUserTypeUnhashed(name string, uType core.UserType) error {
	return db.ChangeUserTypeHashed(auth.HashSHA1(name), uType)
}

func (db *DB) ChangeUserProgressTrackerHashed(hash string, t *core.ProgressTracker) error {
	return db.Update(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
		if err != nil {
			return err
		}
		return user.ChangeProgressTracker(t)
	})
}

// Viewing

func (db *DB) getUser(hash []byte) (*core.User, error) {
	var e *core.User

	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists(hash)
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

func (db *DB) getUserName(hash []byte) (string, error) {
	var e string

	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
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

func (db *DB) GetUserHashed(hash string) (*core.User, error) {
	return db.getUser([]byte(hash))
}

func (db *DB) GetUserUnhashed(name string) (*core.User, error) {
	return db.GetUserHashed(auth.HashSHA1(name))
}

func (db *DB) GetUserNameHashed(hash string) (string, error) {
	return db.getUserName([]byte(hash))
}

func (db *DB) SetSeriesProgressAllRead(usernameHash, sid string) error {
	u, err := db.GetUserHashed(usernameHash)
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
	return db.ChangeUserProgressTrackerHashed(usernameHash, u.ProgressTracker)
}

func (db *DB) SetSeriesProgressAllUnread(usernameHash, sid string) error {
	u, err := db.GetUserHashed(usernameHash)
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
	return db.ChangeUserProgressTrackerHashed(usernameHash, u.ProgressTracker)
}

func (db *DB) SetSeriesEntryProgressRead(usernameHash, sid, eid string) error {
	u, err := db.GetUserHashed(usernameHash)
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
	return db.ChangeUserProgressTrackerHashed(usernameHash, u.ProgressTracker)
}

func (db *DB) SetSeriesEntryProgressUnread(usernameHash, sid, eid string) error {
	u, err := db.GetUserHashed(usernameHash)
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
	return db.ChangeUserProgressTrackerHashed(usernameHash, u.ProgressTracker)
}

func (db *DB) SetSeriesEntryProgressNum(usernameHash, sid, eid string, num int) error {
	u, err := db.GetUserHashed(usernameHash)
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
	return db.ChangeUserProgressTrackerHashed(usernameHash, u.ProgressTracker)
}

func (db *DB) GetUsers(safe bool) ([]core.User, error) {
	users := make([]core.User, 0)
	err := db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		return root.ForEachUser(func(u *core.User) error {
			if safe {
				u.Pass = ""
			}
			users = append(users, *u)
			return nil
		})
	})

	return users, err
}

func (db *DB) HasUsers() bool {
	exists := false

	_ = db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)
		exists = root.HasUsers()
		return nil
	})

	return exists
}

// Processing

func (db *DB) IsUserAdminHashed(hash string) (bool, error) {
	var isAdmin bool

	err :=  db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
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

func (db *DB) IsUserAdminUnhashed(hash string) (bool, error) {
	return db.IsUserAdminHashed(auth.HashSHA1(hash))
}

func (db *DB) AdminCount() (int, error) {
	var count int
	err := db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		c, err := root.AdminCount()
		if err != nil {
			return err
		}
		count = c
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *DB) ValidateLoginHashed(hash, pass string) (bool, error) {
	var valid bool

	err :=  db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUserIfExists([]byte(hash))
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

func (db *DB) ValidateLoginUnhashed(name, pass string) (bool, error) {
	return db.ValidateLoginHashed(auth.HashSHA1(name), pass)
}

func (db *DB) EnsurValidSeriesProgress(sid, eid string, num int) {
	db.Update(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		return root.ForEachUserBucket(func(u *UserBucket) error {
			tracker := u.ProgressTracker()
			if tracker != nil {
				e := tracker.ProgressEntry(sid, eid)
				if e != nil {
					e.Total = num
					u.ChangeProgressTracker(tracker)
				}
			}

			return nil
		})
	})
}
