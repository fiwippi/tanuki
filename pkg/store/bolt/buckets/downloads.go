package buckets

import (
	"encoding/binary"

	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/internal/json"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

type DownloadsBucket struct {
	*bolt.Bucket
}

func (d *DownloadsBucket) NextID() []byte {
	id, err := d.Bucket.NextSequence()
	if err != nil {
		panic(err)
	}

	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, id)
	return b
}

func (d *DownloadsBucket) GetDownloads() []*api.Download {
	downloads := make([]*api.Download, 0)

	// Get the downloads from the db
	d.ForEachDownload(func(_ []byte, dl *api.Download) error {
		downloads = append(downloads, dl)
		return nil
	})

	// Reverse the slice so newest downloads are first
	for i := len(downloads)/2 - 1; i >= 0; i-- {
		opp := len(downloads) - 1 - i
		downloads[i], downloads[opp] = downloads[opp], downloads[i]
	}

	return downloads
}

func (d *DownloadsBucket) DeleteDownload(k []byte) error {
	return d.Delete(k)
}

func (d *DownloadsBucket) AddDownload(dl *api.Download) error {
	return d.Put(d.NextID(), json.Marshal(dl))
}

func (d *DownloadsBucket) ForEachDownload(f func(k []byte, dl *api.Download) error) error {
	return d.ForEach(func(k, v []byte) error {
		return f(k, api.UnmarshalDownload(v))
	})
}
