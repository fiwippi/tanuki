// Package json provides JSON marshalling and unmarshalling for basic types
package json

import "encoding/json"

// Marshal marshals a given object into JSON.
// This must succeed and will panic on fail.
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
