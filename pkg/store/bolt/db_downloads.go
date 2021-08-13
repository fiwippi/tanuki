package bolt

import (
	bolt "go.etcd.io/bbolt"

	"github.com/fiwippi/tanuki/pkg/store/bolt/buckets"
	"github.com/fiwippi/tanuki/pkg/store/bolt/keys"
	"github.com/fiwippi/tanuki/pkg/store/entities/api"
)

func (db *DB) downloadsBucket(tx *bolt.Tx) *buckets.DownloadsBucket {
	return &buckets.DownloadsBucket{Bucket: tx.Bucket(keys.Downloads)}
}

func (db *DB) GetDownloads() []*api.Download {
	var dl []*api.Download
	db.View(func(tx *bolt.Tx) error {
		root := db.downloadsBucket(tx)
		dl = root.GetDownloads()
		return nil
	})
	return dl
}

func (db *DB) GetFailedDownloads() []*api.Download {
	var dls []*api.Download
	err := db.Update(func(tx *bolt.Tx) error {
		root := db.downloadsBucket(tx)
		return root.ForEachDownload(func(k []byte, dl *api.Download) error {
			if dl.Status == api.DownloadFailed {
				dls = append(dls, dl)
				return root.DeleteDownload(k)
			}
			return nil
		})
	})
	if err != nil {
		panic(err)
	}
	return dls
}

func (db *DB) AddDownload(dl *api.Download) error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.downloadsBucket(tx)
		return root.AddDownload(dl)
	})
}

func (db *DB) ClearFinishedDownloads() error {
	return db.Update(func(tx *bolt.Tx) error {
		root := db.downloadsBucket(tx)

		return root.ForEachDownload(func(k []byte, dl *api.Download) error {
			if dl.IsFinished() {
				return root.DeleteDownload(k)
			}
			return nil
		})
	})
}
