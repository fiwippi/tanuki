package proxy

type ProxyRequest struct {
	Endpoint string `json:"endpoint"`
	Query    string `json:"query"`
}
