package human

import "github.com/fiwippi/tanuki/internal/platform/hash"

type User struct {
	UID  string `db:"uid"`
	Name string `db:"name"`
	Pass string `db:"pass"`
	Type Type   `db:"type"`
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
