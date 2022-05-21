package transfer

type DownloadStatus string

const (
	DownloadQueued    DownloadStatus = "Queued"
	DownloadStarted   DownloadStatus = "Started"
	DownloadFinished  DownloadStatus = "Finished"
	DownloadCancelled DownloadStatus = "Cancelled"
	DownloadExists    DownloadStatus = "Already Downloaded"
	DownloadFailed    DownloadStatus = "Failed"
)
