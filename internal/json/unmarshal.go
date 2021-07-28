package json

import (
	"encoding/json"
	"time"
)

func UnmarshalString(data []byte) string {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func UnmarshalInt(data []byte) int {
	var s int
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func UnmarshalTime(data []byte) time.Time {
	var t time.Time
	err := json.Unmarshal(data, &t)
	if err != nil {
		panic(err)
	}
	return t
}
