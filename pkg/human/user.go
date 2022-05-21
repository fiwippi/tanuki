package human

import "github.com/fiwippi/tanuki/internal/hash"

type User struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
	Pass string `json:"pass"`
	Type Type   `json:"type"`
}

// NewUser expects username and unhashed password along with the users permission
func NewUser(name, pass string, t Type) *User {
	return &User{
		Hash: hash.SHA1(name),
		Name: name,
		Pass: hash.SHA256(pass),
		Type: t,
	}
}
