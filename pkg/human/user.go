package human

import "github.com/fiwippi/tanuki/internal/platform/hash"

type User struct {
	UID  string `json:"uid"  db:"uid"`
	Name string `json:"name" db:"name"`
	Pass string `json:"pass" db:"pass"`
	Type Type   `json:"type" db:"type"`
}

// NewUser expects username and unhashed password along with the users permission
func NewUser(name, pass string, t Type) User {
	return User{
		UID:  hash.SHA1(name),
		Name: name,
		Pass: hash.SHA256(pass),
		Type: t,
	}
}
