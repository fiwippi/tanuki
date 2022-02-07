package mangadex

import (
	"encoding/json"
	"fmt"
)

type HomeURLData struct {
	BaseUrl string `json:"baseUrl"`
	Result  string `json:"result"`
	Errors  Errors `json:"errors"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

func (c *Client) GetHomeUrl(cid string) (*HomeURLData, error) {
	resp, err := c.fmtAndSend("GET", fmt.Sprintf("at-home/server/%s", cid), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	var data HomeURLData
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	if len(data.Errors) > 0 {
		return nil, data.Errors[0]
	}

	return &data, nil
}
