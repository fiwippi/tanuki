package storage

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/pkg/user"
)

var ErrUserExist = errors.New("user already exists")
var ErrNotEnoughUsers = errors.New("not enough users in the db")
var ErrNotEnoughAdmins = errors.New("not enough admins in the db")

// Editing

func (s *Store) AddUser(u user.Account, overwrite bool) error {
	fn := func(tx *sqlx.Tx) error {
		if s.hasUser(tx, u.UID) && !overwrite {
			return ErrUserExist
		}
		if s.hasUser(tx, u.UID) && overwrite {
			valid, err := s.validUserTypeChange(tx, u.UID, u.Type)
			if err != nil {
				return err
			}
			if !valid {
				return ErrNotEnoughAdmins
			}
		}

		stmt := `INSERT INTO users (uid, name, pass, type) 
					Values (:uid,:name,:pass,:type)
					ON CONFLICT (uid)
					DO UPDATE SET uid=:uid,name=:name,pass=:pass,type=:type`
		_, err := tx.NamedExec(stmt, u)
		return err
	}

	return s.tx(fn)
}

func (s *Store) DeleteUser(uid string) error {
	fn := func(tx *sqlx.Tx) error {
		var count int
		if err := tx.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
			return err
		}
		if count == 1 {
			return ErrNotEnoughUsers
		}
		_, err := tx.Exec(`DELETE FROM users WHERE uid = ?`, uid)
		return err
	}

	return s.tx(fn)
}

func (s *Store) ChangeUsername(currentUID, newName string) error {
	newUID := hash.SHA1(newName)
	if currentUID == newUID {
		return nil
	}

	fn := func(tx *sqlx.Tx) error {
		if s.hasUser(tx, newUID) {
			return ErrUserExist
		}
		_, err := tx.Exec(`UPDATE users SET uid = ?, name = ? WHERE uid = ?`, newUID, newName, currentUID)
		return err
	}

	return s.tx(fn)
}

func (s *Store) ChangePassword(uid, password string) error {
	_, err := s.pool.Exec(`UPDATE users SET pass = ? WHERE uid = ?`, hash.SHA256(password), uid)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) ChangeType(uid string, t user.Type) error {
	fn := func(tx *sqlx.Tx) error {
		// Check if enough admins
		valid, err := s.validUserTypeChange(tx, uid, t)
		if err != nil {
			return err
		}
		if !valid {
			return ErrNotEnoughAdmins
		}

		// If there will be enough admins then change the user type
		_, err = tx.Exec(`UPDATE users SET type = ? WHERE uid = ?`, t, uid)
		return err
	}

	return s.tx(fn)
}

// Querying

func (s *Store) getUser(tx *sqlx.Tx, uid string) (user.Account, error) {
	var u user.Account
	err := tx.Get(&u, `SELECT * FROM users WHERE uid = ?`, uid)
	if err != nil {
		return user.Account{}, err
	}
	return u, nil
}

func (s *Store) GetUser(uid string) (user.Account, error) {
	var u user.Account
	var err error
	fn := func(tx *sqlx.Tx) error {
		u, err = s.getUser(tx, uid)
		return err
	}

	if err := s.tx(fn); err != nil {
		return user.Account{}, err
	}
	return u, nil
}

func (s *Store) GetUsers() ([]user.Account, error) {
	var u []user.Account
	err := s.pool.Select(&u, `SELECT * FROM users ORDER BY type ASC, ROWID ASC`)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) hasUser(tx *sqlx.Tx, uid string) bool {
	var exists bool
	tx.Get(&exists, "SELECT COUNT(uid) > 0 FROM users WHERE uid = ?", uid)
	return exists
}

func (s *Store) HasUser(uid string) (bool, error) {
	var exists bool
	fn := func(tx *sqlx.Tx) error {
		exists = s.hasUser(tx, uid)
		return nil
	}

	if err := s.tx(fn); err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) HasUsers() (bool, error) {
	var exists bool
	err := s.pool.Get(&exists, "SELECT COUNT(*) > 0 FROM users")
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) validUserTypeChange(tx *sqlx.Tx, uid string, t user.Type) (bool, error) {
	// Get number of admins
	var count int
	tx.Get(&count, `SELECT COUNT(*) FROM users WHERE type = 'admin'`)

	// Ensure 1 admin will exist by the end of the transaction
	u, err := s.getUser(tx, uid)
	if err != nil {
		return false, err
	}
	if u.Type == user.Admin && t != user.Admin {
		count -= 1
	}
	return count > 0, nil
}

// Utility

func (s *Store) IsAdmin(uid string) bool {
	var admin bool
	err := s.pool.Get(&admin, `SELECT type = 'admin' FROM users WHERE uid = ?`, uid)
	if err != nil {
		return false
	}
	return admin
}

func (s *Store) ValidateLogin(name, pass string) bool {
	var valid bool
	err := s.pool.Get(&valid, `SELECT pass = ? FROM users WHERE uid = ?`, hash.SHA256(pass), hash.SHA1(name))
	if err != nil {
		return false
	}
	return valid
}
