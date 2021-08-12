package api

type DownloadStatus string

const (
	Queued    DownloadStatus = "Queued"
	Started   DownloadStatus = "Started"
	Finished  DownloadStatus = "Finished"
	Cancelled DownloadStatus = "Cancelled"
	Exists    DownloadStatus = "Already Downloaded"
	Failed    DownloadStatus = "Failed"
)

func (ds DownloadStatus) Finished() bool {
	return ds != Queued && ds != Started
}
