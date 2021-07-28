package sets

import "encoding/json"

func UnmarshalSet(data []byte) *Set {
	var s Set
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
