package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/auth"
	"github.com/fiwippi/tanuki/pkg/core"
)

type UserBucket struct {
	*bolt.Bucket
}

func (u *UserBucket) Struct() *core.User {
	user := &core.User{
		Hash:     auth.SHA1(u.Name()),
		Name:     u.Name(),
		Pass:     u.Password(),
		Type:     u.Type(),
		Progress: u.Progress(),
	}

	return user
}

func (u *UserBucket) Name() string {
	return core.UnmarshalString(u.Get(keyUsername))
}

func (u *UserBucket) Password() string {
	return core.UnmarshalString(u.Get(keyPassword))
}

func (u *UserBucket) Type() core.UserType {
	return core.UnmarshalUserType(u.Get(keyType))
}

func (u *UserBucket) Progress() *core.CatalogProgress {
	cp := u.Get(keyProgress)
	if cp == nil {
		return core.NewCatalogProgress()
	}
	return core.UnmarshalCatalogProgress(cp)
}

func (u *UserBucket) IsAdmin() bool {
	return u.Type() == core.AdminUser
}

func (u *UserBucket) ValidPassword(unhashedPassword string) bool {
	return u.Password() == auth.SHA256(unhashedPassword)
}

func (u *UserBucket) ChangeName(name string) error {
	return u.Put(keyUsername, core.MarshalJSON(name))
}

func (u *UserBucket) ChangePassword(password string, hash bool) error {
	if hash {
		password = auth.SHA256(password)
	}
	return u.Put(keyPassword, core.MarshalJSON(password))
}

func (u *UserBucket) ChangeType(userType core.UserType) error {
	return u.Put(keyType, core.MarshalJSON(userType))
}

func (u *UserBucket) ChangeProgress(p *core.CatalogProgress) error {
	return u.Put(keyProgress, core.MarshalJSON(p))
}
