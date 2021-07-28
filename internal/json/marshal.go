package json

import "encoding/json"

func Marshal(d interface{}) []byte {
	if d == nil {
		return nil
	}

	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}
