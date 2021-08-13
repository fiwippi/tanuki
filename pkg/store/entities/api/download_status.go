package api

type DownloadStatus string

const (
	DownloadQueued    DownloadStatus = "Queued"
	DownloadStarted   DownloadStatus = "Started"
	DownloadFinished  DownloadStatus = "Finished"
	DownloadCancelled DownloadStatus = "Cancelled"
	DownloadExists    DownloadStatus = "Already Downloaded"
	DownloadFailed    DownloadStatus = "Failed"
)

func (ds DownloadStatus) Finished() bool {
	return ds != DownloadQueued && ds != DownloadStarted
}
