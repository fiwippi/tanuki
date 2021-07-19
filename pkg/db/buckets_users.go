package db

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/core"
)

type UsersBucket struct {
	*bolt.Bucket
}

func (b *UsersBucket) GetUser(usernameHashBytes []byte) *UserBucket {
	bucket := b.Bucket.Bucket(usernameHashBytes)
	if bucket == nil {
		return nil
	}
	return &UserBucket{bucket}
}

func (b *UsersBucket) GetUserIfExists(usernameHashBytes []byte) (*UserBucket, error) {
	if !b.HasUser(usernameHashBytes) {
		return nil, ErrUserNotExist
	}
	return &UserBucket{b.Bucket.Bucket(usernameHashBytes)}, nil
}

func (b *UsersBucket) HasUser(usernameHashBytes []byte) bool {
	return b.Bucket.Bucket(usernameHashBytes) != nil
}

func (b *UsersBucket) HasUsers() bool {
	if k, _ := b.Cursor().First(); k != nil {
		return true
	}
	return false
}

func (b *UsersBucket) AddUser(u *core.User, overwrite bool) error {
	// Check if allowed to add user
	if !overwrite && b.HasUser(u.HashBytes()) {
		return ErrUserExists
	}

	// Write the data
	user, err := b.Bucket.CreateBucketIfNotExists(u.HashBytes())
	if err != nil {
		return err
	}

	err = user.Put(keyUserName, core.MarshalJSON(u.Name))
	if err != nil {
		return err
	}
	err = user.Put(keyUserPassword, core.MarshalJSON(u.Pass))
	if err != nil {
		return err
	}
	err = user.Put(keyUserType, core.MarshalJSON(u.Type))
	if err != nil {
		return err
	}
	err = user.Put(keyUserProgress, core.MarshalJSON(u.ProgressTracker))
	if err != nil {
		return err
	}

	return nil
}

func (b *UsersBucket) DeleteUser(usernameHash []byte) error {
	return b.Bucket.DeleteBucket(usernameHash)
}

func (b *UsersBucket) RenameUser(oldHash, newHash string, newUsername string) error {
	// Old user must exist
	if !b.HasUser([]byte(oldHash)) {
		return ErrUserNotExist
	}

	// New user must not exist
	if b.HasUser([]byte(newHash)) {
		return ErrUserExists
	}

	// Copies data and deletes old user
	err := renameBucket(b.Bucket, b.Bucket, []byte(oldHash), []byte(newHash))
	if err != nil {
		return err
	}

	// Set the new username as well in the user struct
	newUser := b.Bucket.Bucket([]byte(newHash))
	return newUser.Put(keyUserName, core.MarshalJSON(newUsername))
}

func (b *UsersBucket) ForEachUser(f func(u *core.User) error) error {
	return b.Bucket.ForEach(func(k, v []byte) error {
		// Nil means we are accessing a bucket, all users are stored in buckets
		// whilst other data isn't, so we only want buckets
		if v == nil {
			err := f(b.GetUser(k).Struct())
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *UsersBucket) ForEachUserBucket(f func(u *UserBucket) error) error {
	return b.Bucket.ForEach(func(k, v []byte) error {
		// Nil means we are accessing a bucket, all users are stored in buckets
		// whilst other data isn't, so we only want buckets
		if v == nil {
			err := f(b.GetUser(k))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (b *UsersBucket) AdminCount() (int, error) {
	var count int

	err := b.ForEachUser(func(u *core.User) error {
		if u.Type == core.AdminUser {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}



