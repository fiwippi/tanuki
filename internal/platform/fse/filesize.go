package fse

func Filesize(s int64) float64 {
	if s > 0 {
		return float64(s) / 1024 / 1024
	}
	return 0
}
