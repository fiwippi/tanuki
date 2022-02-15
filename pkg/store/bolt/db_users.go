package bolt

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/errors"
	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/pkg/store/bolt/buckets"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

var (
	ErrNotEnoughAdmins  = errors.New("not enough admins in the db")
	ErrProgressNotExist = errors.New("progress does not exist")
)

func (db *DB) usersBucket(tx *bolt.Tx) *buckets.UsersBucket {
	return &buckets.UsersBucket{Bucket: tx.Bucket(keys.Users)}
}

// Editing

func (db *DB) CreateUser(u *users.User) error {
	return db.SaveUser(u, false)
}

func (db *DB) SaveUser(u *users.User, overwrite bool) error {
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
		return root.RenameUser(uid, hash.SHA1(name), name)
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

func (db *DB) ChangeUserType(uid string, uType users.Type) error {
	admins := db.AdminCount()
	return db.Update(func(tx *bolt.Tx) error {
		// Get the user's bucket
		user, err := db.usersBucket(tx).GetUser(uid)
		if err != nil {
			return err
		}

		// Ensure 1 admin will exist by the end of the transaction
		if user.Type() == users.Admin && uType != users.Admin {
			admins -= 1
		}
		if admins == 0 {
			return ErrNotEnoughAdmins
		}

		// Set the password
		return user.ChangeType(uType)
	})
}

func (db *DB) ChangeProgress(uid string, p users.CatalogProgress) error {
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

func (db *DB) GetUser(uid string) (*users.User, error) {
	var e *users.User
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

func (db *DB) GetUsers(safe bool) []users.User {
	users := make([]users.User, 0)
	db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		return root.ForEachUser(func(ub *buckets.UserBucket) error {
			u := ub.Struct()
			if safe {
				u.Pass = ""
			}
			users = append(users, *u)
			return nil
		})
	})

	return users
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

func (db *DB) GetUserProgress(uid string) (users.CatalogProgress, error) {
	var p users.CatalogProgress
	err := db.View(func(tx *bolt.Tx) error {
		root := db.usersBucket(tx)

		user, err := root.GetUser(uid)
		if err != nil {
			return err
		}
		p = user.Progress()

		return nil
	})
	if err != nil {
		return users.CatalogProgress{}, err
	}
	return p, nil
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

func (db *DB) ValidateLogin(name, pass string) bool {
	var valid bool
	db.View(func(tx *bolt.Tx) error {
		user, err := db.usersBucket(tx).GetUser(hash.SHA1(name))
		if err != nil {
			valid = false
		} else {
			valid = user.ValidPassword(pass)
		}

		return nil
	})

	return valid
}
