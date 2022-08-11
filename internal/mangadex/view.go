package mangadex

import (
	"context"
	"encoding/json"
	"net/url"
)

func ViewManga(ctx context.Context, uuid string) (Listing, error) {
	q := url.Values{}
	q.Add("includes[]", "cover_art")

	resp, err := get(ctx, "manga/"+uuid, q)
	if err != nil {
		return Listing{}, err
	}

	r := struct {
		result
		Data listingData `json:"data"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return Listing{}, err
	}

	if r.errored() {
		return Listing{}, r.err()
	}

	return r.Data.makeListing(), nil
}
