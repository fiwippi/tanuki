package cookie

type Action int

const (
	Redirect Action = iota
	Abort
)
