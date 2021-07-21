package db

import (
	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
	"testing"
)

func TestDB_CreateUser(t *testing.T) {
	u1 := core.NewUser("test", "boom", core.StandardUser)

	// Ensure first user can be created successfully
	err := testDb.CreateUser(u1)
	if err != nil {
		t.Error(err)
	}

	// Ensure second user can't be created with same name
	err = testDb.CreateUser(u1)
	if err != ErrUserExists {
		t.Error(err)
	}

	// Ensure third user can be created with different name
	u2 := core.NewUser("different", "", core.StandardUser)
	err = testDb.CreateUser(u2)
	if err != nil {
		t.Error(err)
	}

	// Ensure third user can be viewed
	_, err = testDb.GetUser(auth.HashSHA1(u2.Name))
	if err != nil {
		t.Error(err)
	}

	// Delete third user and make sure it cant be viewed
	err = testDb.DeleteUser(auth.HashSHA1(u2.Name))
	if err != nil {
		t.Error(err)
	}

	_, err = testDb.GetUser(auth.HashSHA1(u2.Name))
	if err != ErrUserNotExist {
		t.Error(err)
	}

	// Fake login validation for the first user
	valid, err := testDb.ValidateLogin("test", "boom")
	if err != nil {
		t.Error(err)
	} else if !valid {
		t.Error("user should be valid but isn't")
	}

	//
	testDb.DeleteUser(auth.HashSHA1(u1.Name))
	testDb.DeleteUser(auth.HashSHA1(u2.Name))
}

func TestDB_GetUsers(t *testing.T) {
	u1 := core.NewUser("a", "hahaha", core.AdminUser)
	u2 := core.NewUser("b", "jajaja", core.StandardUser)
	u3 := core.NewUser("c", "kekeke", core.StandardUser)

	// Ensure first user can be created successfully
	if err := testDb.CreateUser(u1); err != nil {
		t.Error(err)
	}
	if err := testDb.CreateUser(u2); err != nil {
		t.Error(err)
	}
	if err := testDb.CreateUser(u3); err != nil {
		t.Error(err)
	}

	users, err := testDb.GetUsers(true)
	if err != nil {
		t.Error(err)
	}

	if len(users) == 0 {
		t.Error("should be users in the database")
	}

	if users[0].Pass != "" {
		t.Error("passwords should be sanitised")
	}

	testDb.DeleteUser(auth.HashSHA1(u1.Name))
	testDb.DeleteUser(auth.HashSHA1(u2.Name))
	testDb.DeleteUser(auth.HashSHA1(u3.Name))
}

func TestDB_ChangeUsername(t *testing.T) {
	u1 := core.NewUser("a", "hahaha", core.AdminUser)
	u2 := core.NewUser("b", "jajaja", core.StandardUser)

	// Ensure first user can be created successfully
	if err := testDb.CreateUser(u1); err != nil {
		t.Error(err)
	}
	if err := testDb.CreateUser(u2); err != nil {
		t.Error(err)
	}

	// Check a can be renamed to c, meaning c exists and a does not exist
	err := testDb.ChangeUsername(u1.Hash, "c")
	if err != nil {
		t.Error("change username should be successful, err:", err)
	}

	u, err := testDb.GetUser(auth.HashSHA1(u1.Name))
	if err != ErrUserNotExist {
		t.Error("user 'a' should not exist:", u, err)
	}

	u, err = testDb.GetUser(auth.HashSHA1("c"))
	if err != nil {
		t.Error("user 'c' should exist:", u, err)
	}

	testDb.DeleteUser(auth.HashSHA1(u1.Name))
	testDb.DeleteUser(auth.HashSHA1(u2.Name))
	testDb.DeleteUser(auth.HashSHA1("c"))
}

func TestDB_ChangeType(t *testing.T) {
	u1 := core.NewUser("a", "hahaha", core.AdminUser)
	u2 := core.NewUser("b", "jajaja", core.AdminUser)

	if err := testDb.CreateUser(u1); err != nil {
		t.Error(err)
	}
	if err := testDb.CreateUser(u2); err != nil {
		t.Error(err)
	}

	// 'a' should be able to become a standard user without any problems
	err := testDb.ChangeUserType(auth.HashSHA1("a"), core.StandardUser)
	if err != nil {
		t.Error("change type should be successful", err)
	}
	a, err := testDb.GetUser(auth.HashSHA1(u1.Name))
	if err != nil {
		t.Error(err)
	}
	if a.Type != core.StandardUser {
		t.Error("type did not change from admin to standard user")
	}

	// changing 'b' from admin to standard user should result in ErrNotEnoughAdmin
	err = testDb.ChangeUserType(auth.HashSHA1("b"), core.StandardUser)
	if err != ErrNotEnoughAdmins {
		t.Error("change type raised wrong error", err)
	}

	testDb.DeleteUser(auth.HashSHA1(u1.Name))
	testDb.DeleteUser(auth.HashSHA1(u2.Name))
}

func TestDB_ChangePassword(t *testing.T) {
	u1 := core.NewUser("a", "hahaha", core.AdminUser)

	// Ensure first user can be created successfully
	if err := testDb.CreateUser(u1); err != nil {
		t.Error(err)
	}

	// Check a can be renamed to c, meaning c exists and a does not exist
	newPassHash := auth.HashSHA256("kekeke")
	err := testDb.ChangePassword(auth.HashSHA1(u1.Name), "kekeke")
	if err != nil {
		t.Error("password should have changed successfully", err)
	}

	u, err := testDb.GetUser(auth.HashSHA1(u1.Name))
	if err != nil {
		t.Error(err)
	}
	if u.Pass != newPassHash {
		t.Error("password should have changed to the new hash")
	}

	testDb.DeleteUser(auth.HashSHA1(u1.Name))
}
