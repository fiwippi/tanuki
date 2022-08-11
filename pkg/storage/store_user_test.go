package storage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fiwippi/tanuki/internal/platform/hash"
	"github.com/fiwippi/tanuki/pkg/human"
)

// Editing

func TestStore_AddUser(t *testing.T) {
	s := mustOpenStoreMem(t)

	uAdmin := human.NewUser("a", "", human.Admin)
	uStandard := human.NewUser("a", "", human.Standard)
	require.Nil(t, s.AddUser(uAdmin, true))

	// Can't save if overwrite is false
	err := s.AddUser(uAdmin, false)
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrUserExist)

	// Can save if overwrite is true, verify this by seeing the user type has changed
	require.Nil(t, s.AddUser(uStandard, true))
	u, err := s.GetUser(uAdmin.UID)
	require.Nil(t, err)
	require.Equal(t, human.Standard, u.Type)

	// If there is only one user left in the DB we can't overwrite their
	// save with a standard user since there would be no admins
	require.Nil(t, s.DeleteUser(uStandard.UID))
	err = s.AddUser(human.NewUser("default", "", human.Standard), true)
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrNotEnoughAdmins)
	mustCloseStore(t, s)
}

func TestStore_DeleteUser(t *testing.T) {
	s := mustOpenStoreMem(t)

	// Can't delete if only one user in DB
	err := s.DeleteUser(defaultUID)
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrNotEnoughUsers)

	// Can delete if at least two users in DB
	require.Nil(t, s.AddUser(human.NewUser("a", "", human.Admin), true))
	require.Nil(t, s.DeleteUser(defaultUID))

	mustCloseStore(t, s)
}

func TestStore_ChangeUsername(t *testing.T) {
	s := mustOpenStoreMem(t)

	u1 := human.NewUser("a", "", human.Admin)
	u2 := human.NewUser("b", "", human.Admin)
	require.Nil(t, s.AddUser(u1, true))
	require.Nil(t, s.AddUser(u2, true))
	require.Nil(t, s.DeleteUser(defaultUID))

	// Check 'a' being renamed to itself is ignored
	err := s.ChangeUsername(u1.UID, "a")
	require.Nil(t, err)

	// Check 'a' cannot be renamed to 'b' since it already exists
	err = s.ChangeUsername(u1.UID, "b")
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrUserExist)

	// Check 'a' can be renamed to 'c', meaning 'c' exists and 'a' does not exist
	err = s.ChangeUsername(u1.UID, "c")
	require.Nil(t, err)

	has, err := s.HasUser(hash.SHA1("a"))
	require.Nil(t, err)
	require.False(t, has)
	has, err = s.HasUser(hash.SHA1("c"))
	require.Nil(t, err)
	require.True(t, has)

	mustCloseStore(t, s)
}

func TestStore_ChangePassword(t *testing.T) {
	s := mustOpenStoreMem(t)

	err := s.ChangePassword(defaultUID, "a")
	require.Nil(t, err)

	u, err := s.GetUser(defaultUID)
	require.Nil(t, err)
	require.Equal(t, hash.SHA256("a"), u.Pass)

	mustCloseStore(t, s)
}

func TestStore_ChangeType(t *testing.T) {
	s := mustOpenStoreMem(t)

	u1 := human.NewUser("a", "", human.Admin)
	u2 := human.NewUser("b", "", human.Admin)
	require.Nil(t, s.AddUser(u1, true))
	require.Nil(t, s.AddUser(u2, true))
	require.Nil(t, s.DeleteUser(defaultUID))

	// Change to standard when there will still
	// exist one admin afterwards should succeed
	err := s.ChangeType(u1.UID, human.Standard)
	require.Nil(t, err)
	u, err := s.GetUser(u1.UID)
	require.Nil(t, err)
	require.Equal(t, human.Standard, u.Type)

	// Change to standard when there will be
	// no admins afterwards should not succeed
	err = s.ChangeType(u2.UID, human.Standard)
	require.NotNil(t, err)
	require.ErrorIs(t, err, ErrNotEnoughAdmins)

	mustCloseStore(t, s)
}

// Querying

func TestStore_GetUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	u, err := s.GetUser(defaultUID)
	require.Nil(t, err)
	require.Equal(t, defaultUID, u.UID)
	require.Equal(t, "default", u.Name)
	require.Equal(t, human.Admin, u.Type)
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
		require.Nil(t, s.AddUser(u, true))
	}
	require.Nil(t, s.DeleteUser(defaultUID))

	dbUsers, err := s.GetUsers()
	require.Nil(t, err)

	// Check users returned in order of insertion but
	// they should also first be separated into Admin
	// or Standard, so the order should be a, c, b, d
	require.Equal(t, users[0], dbUsers[0]) // a
	require.Equal(t, users[1], dbUsers[2]) // c
	require.Equal(t, users[2], dbUsers[1]) // b
	require.Equal(t, users[3], dbUsers[3]) // d

	mustCloseStore(t, s)
}

func TestStore_HasUser(t *testing.T) {
	s := mustOpenStoreMem(t)
	has, err := s.HasUser(defaultUID)
	require.Nil(t, err)
	require.True(t, has)
	mustCloseStore(t, s)
}

func TestStore_HasUsers(t *testing.T) {
	s := mustOpenStoreMem(t)
	has, err := s.HasUsers()
	require.Nil(t, err)
	require.True(t, has)
	mustCloseStore(t, s)
}

// Utility

func TestStore_IsAdmin(t *testing.T) {
	s := mustOpenStoreMem(t)
	isAdmin := s.IsAdmin(defaultUID)
	require.True(t, isAdmin)
	mustCloseStore(t, s)
}

func TestStore_ValidateLogin(t *testing.T) {
	s := mustOpenStoreMem(t)
	require.Nil(t, s.AddUser(human.NewUser("a", "b", human.Admin), true))
	valid := s.ValidateLogin("a", "b")
	require.True(t, valid)
	mustCloseStore(t, s)
}
