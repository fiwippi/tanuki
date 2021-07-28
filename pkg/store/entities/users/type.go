package users

import "encoding/json"

type Type string

const (
	Admin    Type = "admin"
	Standard Type = "standard"
)

func UnmarshalType(data []byte) Type {
	var s Type
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}
