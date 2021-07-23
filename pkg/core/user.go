package core

import (
	"github.com/fiwippi/tanuki/pkg/auth"
)

type User struct {
	Hash     string           `json:"hash"`
	Name     string           `json:"name"`
	Pass     string           `json:"pass"`
	Type     UserType         `json:"type"`
	Progress *CatalogProgress `json:"progress"`
}

// NewUser expects username and unhashed password along with the users permission
func NewUser(name, pass string, t UserType) *User {
	return &User{
		Hash:     auth.SHA1(name),
		Name:     name,
		Pass:     auth.SHA256(pass),
		Type:     t,
		Progress: NewCatalogProgress(),
	}
}
