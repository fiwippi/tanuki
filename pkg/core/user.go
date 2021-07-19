package core

import (
	"github.com/fiwippi/tanuki/pkg/auth"
)

type User struct {
	Hash            string           `json:"hash,omitempty"`
	Name            string           `json:"name"`
	Pass            string           `json:"pass"`
	Type            UserType         `json:"type"`
	ProgressTracker *ProgressTracker `json:"progress_tracker"`
}

// NewUser expects username and unhashed password along with the users permission
func NewUser(name, pass string, uType UserType) *User {
	return &User{
		Name:            name,
		Pass:            auth.HashSHA256(pass),
		Type:            uType,
		ProgressTracker: NewProgressTracker(),
	}
}

func (u *User) HashBytes() []byte {
	return []byte(u.HashString())
}

func (u *User) HashString() string {
	if u.Hash == "" {
		u.Hash = auth.HashSHA1(u.Name)
	}
	return u.Hash
}
