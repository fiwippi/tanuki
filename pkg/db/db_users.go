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
		return root.RenameUser(uid, auth.SHA1(name), name)
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

func (db *DB) ChangeProgress(uid string, p *core.CatalogProgress) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Get the users bucket
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}

		// Set the password
		return user.ChangeProgress(p)
	})
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

func (db *DB) AdminCount() int {
	var count int
	db.View(func(tx *bolt.Tx) error {
		count = db.usersBucket(tx).AdminCount()
		return nil
	})
	return count
}

func (db *DB) GetProgress(uid string) (*core.CatalogProgress, error) {
	var p *core.CatalogProgress
	err := db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		user, err := root.GetUser(uid)
		if err != nil {
			return err
		}

		progress := user.Progress()
		if progress == nil {
			return ErrProgressNotExist
		}
		p = progress

		return nil
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Processing

func (db *DB) IsAdmin(uid string) (bool, error) {
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

func (db *DB) ValidateLogin(name, pass string) (bool, error) {
	var valid bool
	err := db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(auth.SHA1(name))
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
