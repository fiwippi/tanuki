package core

import (
	"encoding/json"
	"time"

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

func UnmarshalCatalogProgress(data []byte) *CatalogProgress {
	var p CatalogProgress
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

func UnmarshalOrder(data []byte) int {
	if data == nil {
		return -1
	}

	var s int
	err := json.Unmarshal(data, &s)
	if err != nil {
		panic(err)
	}
	return s
}

func UnmarshalPage(data []byte) *Page {
	if data == nil {
		return nil
	}

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

func UnmarshalTime(data []byte) time.Time {
	var t time.Time
	err := json.Unmarshal(data, &t)
	if err != nil {
		panic(err)
	}
	return t
}
