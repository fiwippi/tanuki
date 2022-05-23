package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/pkg/human"
)

// Editing

func TestStore_CreateUser(t *testing.T) {
	s := mustOpenStoreMem(t)

	uAdmin := human.NewUser("a", "", human.Admin)
	uStandard := human.NewUser("a", "", human.Standard)
	assert.Nil(t, s.CreateUser(uAdmin))

	// Can overwrite users which exist, seen by change of type
	assert.Nil(t, s.CreateUser(uStandard))
	u, err := s.GetUser(uAdmin.UID)
	assert.Nil(t, err)
	assert.Equal(t, human.Standard, u.Type)

	// If there is only one user left in the DB we can't overwrite their
	// save with a standard user since there would be no admins
	assert.Nil(t, s.DeleteUser(uStandard.UID))
	err = s.CreateUser(human.NewUser("default", "", human.Standard))
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotEnoughAdmins)
	mustCloseStore(t, s)
}

func TestStore_SaveUser(t *testing.T) {
	s := mustOpenStoreMem(t)

	uAdmin := human.NewUser("a", "", human.Admin)
	uStandard := human.NewUser("a", "", human.Standard)
	assert.Nil(t, s.CreateUser(uAdmin))

	// Can't save if overwrite is false
	err := s.SaveUser(uAdmin, false)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrUserExists)

	// Can save if overwrite is true, verify this by seeing the user type has changed
	assert.Nil(t, s.SaveUser(uStandard, true))
	u, err := s.GetUser(uAdmin.UID)
	assert.Nil(t, err)
	assert.Equal(t, human.Standard, u.Type)

	// If there is only one user left in the DB we can't overwrite their
	// save with a standard user since there would be no admins
	assert.Nil(t, s.DeleteUser(uStandard.UID))
	err = s.SaveUser(human.NewUser("default", "", human.Standard), true)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotEnoughAdmins)
	mustCloseStore(t, s)
}

func TestStore_DeleteUser(t *testing.T) {
	s := mustOpenStoreMem(t)

	// Can't delete if only one user in DB
	err := s.DeleteUser(defaultUID)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotEnoughUsers)

	// Can delete if at least two users in DB
	assert.Nil(t, s.CreateUser(human.NewUser("a", "", human.Admin)))
	assert.Nil(t, s.DeleteUser(defaultUID))

	mustCloseStore(t, s)
}

func TestStore_ChangeUsername(t *testing.T) {
	s := mustOpenStoreMem(t)

	u1 := human.NewUser("a", "", human.Admin)
	u2 := human.NewUser("b", "", human.Admin)
	assert.Nil(t, s.CreateUser(u1))
	assert.Nil(t, s.CreateUser(u2))
	assert.Nil(t, s.DeleteUser(defaultUID))

	// Check 'a' being renamed to itself is ignored
	err := s.ChangeUsername(u1.UID, "a")
	assert.Nil(t, err)

	// Check 'a' cannot be renamed to 'b' since it already exists
	err = s.ChangeUsername(u1.UID, "b")
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrUserExists)

	// Check 'a' can be renamed to 'c', meaning 'c' exists and 'a' does not exist
	err = s.ChangeUsername(u1.UID, "c")
	assert.Nil(t, err)

	has, err := s.HasUser(hash.SHA1("a"))
	assert.Nil(t, err)
	assert.False(t, has)
	has, err = s.HasUser(hash.SHA1("c"))
	assert.Nil(t, err)
	assert.True(t, has)

	mustCloseStore(t, s)
}

func TestStore_ChangePassword(t *testing.T) {
	s := mustOpenStoreMem(t)

	err := s.ChangePassword(defaultUID, "a")
	assert.Nil(t, err)

	u, err := s.GetUser(defaultUID)
	assert.Nil(t, err)
	assert.Equal(t, hash.SHA256("a"), u.Pass)

	mustCloseStore(t, s)
}

func TestStore_ChangeType(t *testing.T) {
	s := mustOpenStoreMem(t)

	u1 := human.NewUser("a", "", human.Admin)
	u2 := human.NewUser("b", "", human.Admin)
	assert.Nil(t, s.CreateUser(u1))
	assert.Nil(t, s.CreateUser(u2))
	assert.Nil(t, s.DeleteUser(defaultUID))

	// Change to standard when there will still
	// exist one admin afterwards should succeed
	err := s.ChangeType(u1.UID, human.Standard)
	assert.Nil(t, err)
	u, err := s.GetUser(u1.UID)
	assert.Nil(t, err)
	assert.Equal(t, human.Standard, u.Type)

	// Change to standard when there will be
	// no admins afterwards should not succeed
	err = s.ChangeType(u2.UID, human.Standard)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, ErrNotEnoughAdmins)

	mustCloseStore(t, s)
}

// Querying

func TestStore_GetUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	u, err := s.GetUser(defaultUID)
	assert.Nil(t, err)
	assert.Equal(t, defaultUID, u.UID)
	assert.Equal(t, "default", u.Name)
	assert.Equal(t, human.Admin, u.Type)
	mustCloseStore(t, s)
}

func TestStore_GetUsers(t *testing.T) {
	s := mustOpenStoreMem(t)

	users := []human.User{
		human.NewUser("a", "a", human.Admin),
		human.NewUser("b", "b", human.Standard),
		human.NewUser("c", "c", human.Admin),
		human.NewUser("d", "d", human.Standard),
	}
	for _, u := range users {
		assert.Nil(t, s.CreateUser(u))
	}
	assert.Nil(t, s.DeleteUser(defaultUID))

	dbUsers, err := s.GetUsers()
	assert.Nil(t, err)

	// Check users returned in order of insertion but
	// they should also first be separated into Admin
	// or Standard, so the order should be a, c, b, d
	assert.Equal(t, users[0], dbUsers[0]) // a
	assert.Equal(t, users[1], dbUsers[2]) // c
	assert.Equal(t, users[2], dbUsers[1]) // b
	assert.Equal(t, users[3], dbUsers[3]) // d

	mustCloseStore(t, s)
}

func TestStore_HasUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	has, err := s.HasUser(defaultUID)
	assert.Nil(t, err)
	assert.True(t, has)
	mustCloseStore(t, s)
}

func TestStore_HasUsers(t *testing.T) {
	s := mustOpenStoreMem(t)
	has, err := s.HasUsers()
	assert.Nil(t, err)
	assert.True(t, has)
	mustCloseStore(t, s)
}

// Utility

func TestStore_IsAdmin(t *testing.T) {
	s := mustOpenStoreMem(t)
	isAdmin, err := s.IsAdmin(defaultUID)
	assert.Nil(t, err)
	assert.True(t, isAdmin)
	mustCloseStore(t, s)
}

func TestStore_ValidateLogin(t *testing.T) {
	s := mustOpenStoreMem(t)
	assert.Nil(t, s.CreateUser(human.NewUser("a", "b", human.Admin)))
	valid, err := s.ValidateLogin("a", "b")
	assert.Nil(t, err)
	assert.True(t, valid)
	mustCloseStore(t, s)
}
