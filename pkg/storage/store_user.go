package storage

import (
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/pkg/human"
)

var ErrUserExists = errors.New("user already exists")
var ErrNotEnoughUsers = errors.New("not enough users in the db")
var ErrNotEnoughAdmins = errors.New("not enough admins in the db")

// Editing

func (s *Store) CreateUser(u human.User) error {
	return s.SaveUser(u, true)
}

func (s *Store) SaveUser(u human.User, overwrite bool) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if s.hasUser(tx, u.UID) && !overwrite {
		return ErrUserExists
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

	_, err = tx.NamedExec(`REPLACE INTO users (uid, name, pass, type) Values (:uid,:name,:pass,:type)`, u)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) DeleteUser(uid string) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	if err = tx.Get(&count, "SELECT COUNT(*) FROM users"); err != nil {
		return err
	}
	if count == 1 {
		return ErrNotEnoughUsers
	}
	if _, err = tx.Exec(`DELETE FROM users WHERE uid = ?`, uid); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ChangeUsername(currentUID, newName string) error {
	newUID := hash.SHA1(newName)
	if currentUID == newUID {
		return nil
	}

	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if s.hasUser(tx, newUID) {
		return ErrUserExists
	}
	_, err = tx.Exec(`UPDATE users SET uid = ?, name = ? WHERE uid = ?`, newUID, newName, currentUID)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *Store) ChangePassword(uid, password string) error {
	_, err := s.pool.Exec(`UPDATE users SET pass = ? WHERE uid = ?`, hash.SHA256(password), uid)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) ChangeType(uid string, t human.Type) error {
	tx, err := s.pool.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Querying

func (s *Store) getUser(tx *sqlx.Tx, uid string) (human.User, error) {
	var u human.User
	err := tx.Get(&u, `SELECT * FROM users WHERE uid = ?`, uid)
	if err != nil {
		return human.User{}, err
	}
	return u, nil
}

func (s *Store) GetUser(uid string) (human.User, error) {
	tx, err := s.pool.Beginx()
	if err != nil {
		return human.User{}, err
	}
	defer tx.Rollback()

	u, err := s.getUser(tx, uid)
	if err != nil {
		return human.User{}, err
	}

	if err := tx.Commit(); err != nil {
		return human.User{}, err
	}
	return u, nil
}

func (s *Store) GetUsers() ([]human.User, error) {
	var u []human.User
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
	tx, err := s.pool.Beginx()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	valid := s.hasUser(tx, uid)

	if err := tx.Commit(); err != nil {
		return false, err
	}
	return valid, nil
}

func (s *Store) HasUsers() (bool, error) {
	var exists bool
	err := s.pool.Get(&exists, "SELECT COUNT(*) > 0 FROM users")
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Store) validUserTypeChange(tx *sqlx.Tx, uid string, t human.Type) (bool, error) {
	// Get number of admins
	var count int
	tx.Get(&count, `SELECT COUNT(*) FROM users WHERE type = 'admin'`)

	// Ensure 1 admin will exist by the end of the transaction
	u, err := s.getUser(tx, uid)
	if err != nil {
		return false, err
	}
	if u.Type == human.Admin && t != human.Admin {
		count -= 1
	}
	return count > 0, nil
}

// Utility

func (s *Store) IsAdmin(uid string) (bool, error) {
	var admin bool
	err := s.pool.Get(&admin, `SELECT type = 'admin' FROM users WHERE uid = ?`, uid)
	if err != nil {
		return false, err
	}
	return admin, nil
}

func (s *Store) ValidateLogin(name, pass string) (bool, error) {
	var valid bool
	err := s.pool.Get(&valid, `SELECT pass = ? FROM users WHERE uid = ?`, hash.SHA256(pass), hash.SHA1(name))
	if err != nil {
		return false, err
	}
	return valid, nil
}
