package json

import (
	"encoding/json"
	"time"
)

// UnmarshalString unmarshalls a string from JSON.
// Will panic on fail
func UnmarshalString(data []byte) string {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

// UnmarshalInt unmarshalls an int from JSON.
// Will panic on fail
func UnmarshalInt(data []byte) int {
	var s int
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

// UnmarshalTime unmarshalls a time.Time from JSON.
// Will panic on fail
func UnmarshalTime(data []byte) time.Time {
	var t time.Time
	err := json.Unmarshal(data, &t)
	if err != nil {
		panic(err)
	}
	return t
}
