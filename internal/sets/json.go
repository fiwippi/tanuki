package sets

import "encoding/json"

func UnmarshalSet(data []byte) *Set {
	if data == nil {
		return nil
	}

	var s Set
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}
