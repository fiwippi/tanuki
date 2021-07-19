package core

import (
	"encoding/json"

	"github.com/fiwippi/tanuki/internal/sets"
)

func MarshalJSON(d interface{}) []byte {
	if d == nil {
		return nil
	}

	b, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}
	return b
}

func UnmarshalCover(data []byte) *Cover {
	if data == nil {
		return nil
	}

	var s Cover
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}

func UnmarshalArchive(data []byte) *Archive {
	var s Archive
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}

func UnmarshalProgressTracker(data []byte) *ProgressTracker {
	var p ProgressTracker
	err := json.Unmarshal(data, &p)
	if err != nil {
		panic(err)
	}
	return &p
}

func UnmarshalSet(data []byte) *sets.Set {
	if data == nil {
		return nil
	}

	var s sets.Set
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}

func UnmarshalString(data []byte) string {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func UnmarshalPage(data []byte) *Page {
	var s Page
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return &s
}

func UnmarshalUserType(data []byte) UserType {
	var p UserType
	err := json.Unmarshal(data, &p)
	if err != nil {
		panic(err)
	}
	return p
}

func UnmarshalSeriesMetadata(data []byte) *SeriesMetadata {
	if data == nil {
		return nil
	}

	var p SeriesMetadata
	err := json.Unmarshal(data, &p)
	if err != nil {
		panic(err)
	}
	return &p
}

func UnmarshalEntryMetadata(data []byte) *EntryMetadata {
	if data == nil {
		return nil
	}

	var p EntryMetadata
	err := json.Unmarshal(data, &p)
	if err != nil {
		panic(err)
	}
	return &p
}

