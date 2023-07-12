package mangadex

import (
	"errors"
	"fmt"
)

type result struct {
	Result string `json:"result"`
	Errors []struct {
		ID     string `json:"id"`
		Status int    `json:"status"`
		Title  string `json:"title"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

func (r result) errored() bool {
	return r.Result == "error"
}

func (r result) err() error {
	msg := "mangadex error"
	for _, e := range r.Errors {
		msg += fmt.Sprintf(": %s - %s", e.Title, e.Detail)
	}

	return errors.New(msg)
}
