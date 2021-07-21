package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/core"
)

type UsersBucket struct {
	*bolt.Bucket
}

func (b *UsersBucket) GetUser(uid string) (*UserBucket, error) {
	bucket := b.Bucket.Bucket([]byte(uid))
	if bucket == nil {
		return nil, ErrUserNotExist
	}
	return &UserBucket{bucket}, nil
}

func (b *UsersBucket) HasUser(uid string) bool {
	_, err := b.GetUser(uid)
	return err != ErrUserNotExist
}

func (b *UsersBucket) HasUsers() bool {
	if k, _ := b.Cursor().First(); k != nil {
		return true
	}
	return false
}

func (b *UsersBucket) AddUser(u *core.User, overwrite bool) error {
	// Check if allowed to add user
	if !overwrite && b.HasUser(u.HashString()) {
		return ErrUserExists
	}

	// Write the data
	temp, err := b.Bucket.CreateBucketIfNotExists(u.HashBytes())
	if err != nil {
		return err
	}
	user := &UserBucket{temp}

	err = user.ChangeName(u.Name)
	if err != nil {
		return err
	}
	err = user.ChangePassword(u.Pass, false)
	if err != nil {
		return err
	}
	err = user.ChangeType(u.Type)
	if err != nil {
		return err
	}
	err = user.ChangeProgressTracker(u.ProgressTracker)
	if err != nil {
		return err
	}

	return nil
}

func (b *UsersBucket) DeleteUser(uid string) error {
	return b.Bucket.DeleteBucket([]byte(uid))
}

func (b *UsersBucket) RenameUser(oldUid, newUid string, newUsername string) error {
	// Old user must exist
	if !b.HasUser(oldUid) {
		return ErrUserNotExist
	}

	// New user must not exist
	if b.HasUser(newUid) {
		return ErrUserExists
	}

	// Copies data and deletes old user
	err := renameBucket(b.Bucket, b.Bucket, []byte(oldUid), []byte(newUid))
	if err != nil {
		return err
	}

	// Set the new username as well in the user struct
	newUser := b.Bucket.Bucket([]byte(newUid))
	return newUser.Put(keyUserName, core.MarshalJSON(newUsername))
}

func (b *UsersBucket) ForEachUser(f func(ub *UserBucket) error) error {
	return b.Bucket.ForEach(func(k, v []byte) error {
		// Nil means we are accessing a bucket, all users are stored in buckets
		// whilst other data isn't, so we only want buckets
		if v == nil {
			bucket, _ := b.GetUser(string(k))
			err := f(bucket)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *UsersBucket) AdminCount() int {
	var count int
	b.ForEachUser(func(ub *UserBucket) error {
		if ub.Type() == core.AdminUser {
			count++
		}
		return nil
	})

	return count
}
