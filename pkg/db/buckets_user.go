package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

var (
	keyUserName     = []byte("username")
	keyUserPassword = []byte("password")
	keyUserType     = []byte("type")
	keyUserProgress = []byte("progress")
)

type UserBucket struct {
	*bolt.Bucket
}

func (u *UserBucket) Struct() *core.User {
	user := &core.User{
		Name:            u.Name(),
		Pass:            u.Password(),
		Type:            u.Type(),
		ProgressTracker: u.ProgressTracker(),
	}
	user.HashString() // Generates the hash

	return user
}

func (u *UserBucket) Name() string {
	return core.UnmarshalString(u.Get(keyUserName))
}

func (u *UserBucket) Password() string {
	return core.UnmarshalString(u.Get(keyUserPassword))
}

func (u *UserBucket) Type() core.UserType {
	return core.UnmarshalUserType(u.Get(keyUserType))
}

func (u *UserBucket) ProgressTracker() *core.ProgressTracker {
	return core.UnmarshalProgressTracker(u.Get(keyUserProgress))
}

func (u *UserBucket) IsAdmin() bool {
	return u.Type() == core.AdminUser
}

func (u *UserBucket) ValidPassword(unhashedPassword string) bool {
	return u.Password() == auth.HashSHA256(unhashedPassword)
}

func (u *UserBucket) ChangePassword(unhashedPassword string) error {
	return u.Put(keyUserPassword, core.MarshalJSON(auth.HashSHA256(unhashedPassword)))
}

func (u *UserBucket) ChangeType(userType core.UserType) error {
	return u.Put(keyUserType, core.MarshalJSON(userType))
}

func (u *UserBucket) ChangeProgressTracker(t *core.ProgressTracker) error {
	return u.Put(keyUserProgress, core.MarshalJSON(t))
}
