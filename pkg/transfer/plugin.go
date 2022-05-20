package transfer

import "context"

type Plugin interface {
	Search(ctx context.Context, term string, limit int) ([]Listing, error)
	ViewChapters(ctx context.Context, l Listing) ([]Chapter, error)
	Download(ctx context.Context, ch Chapter, folder string) error
}
