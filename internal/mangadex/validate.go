package mangadex

import (
	"context"
	"encoding/json"
)

// TODO do validation when the user tries to add a uuid and not the system

func ValidateManga(ctx context.Context, uuid string) (bool, error) {
	resp, err := get(ctx, "manga/"+uuid, nil)
	if err != nil {
		return false, err
	}

	var r result
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return false, err
	}
	if r.errored() {
		return false, r.err()
	}

	return r.Result == "ok", nil
}
