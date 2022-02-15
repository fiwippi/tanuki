package buckets

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/hash"
	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/users"
)

type UserBucket struct {
	*bolt.Bucket
}

func (u *UserBucket) Struct() *users.User {
	user := &users.User{
		Hash:     hash.SHA1(u.Name()),
		Name:     u.Name(),
		Pass:     u.Password(),
		Type:     u.Type(),
		Progress: u.Progress(),
	}

	return user
}

func (u *UserBucket) Name() string {
	return json.UnmarshalString(u.Get(keys.Username))
}

func (u *UserBucket) Password() string {
	return json.UnmarshalString(u.Get(keys.Password))
}

func (u *UserBucket) Type() users.Type {
	return users.UnmarshalType(u.Get(keys.Type))
}

func (u *UserBucket) Progress() users.CatalogProgress {
	cp := u.Get(keys.Progress)
	if cp == nil {
		return users.NewCatalogProgress()
	}
	return users.UnmarshalCatalogProgress(cp)
}

func (u *UserBucket) IsAdmin() bool {
	return u.Type() == users.Admin
}

func (u *UserBucket) ValidPassword(unhashedPassword string) bool {
	return u.Password() == hash.SHA256(unhashedPassword)
}

func (u *UserBucket) ChangeName(name string) error {
	return u.Put(keys.Username, json.Marshal(name))
}

func (u *UserBucket) ChangePassword(password string, shouldHash bool) error {
	if shouldHash {
		password = hash.SHA256(password)
	}
	return u.Put(keys.Password, json.Marshal(password))
}

func (u *UserBucket) ChangeType(userType users.Type) error {
	return u.Put(keys.Type, json.Marshal(userType))
}

func (u *UserBucket) ChangeProgress(p users.CatalogProgress) error {
	return u.Put(keys.Progress, json.Marshal(p))
}
