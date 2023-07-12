package user

import "github.com/fiwippi/tanuki/internal/hash"

type Account struct {
	UID  string `json:"uid"  db:"uid"`
	Name string `json:"name" db:"name"`
	Pass string `json:"pass" db:"pass"`
	Type Type   `json:"type" db:"type"`
}

// NewAccount expects username and unhashed password along with the users permission
func NewAccount(name, pass string, t Type) Account {
	return Account{
		UID:  hash.SHA1(name),
		Name: name,
		Pass: hash.SHA256(pass),
		Type: t,
	}
}
