package mangadex

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetHomeUrl(cid string) (string, error) {
	resp, err := c.fmtAndSend("GET", fmt.Sprintf("at-home/server/%s", cid), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)

	data := struct {
		BaseUrl string `json:"baseUrl"`
		Result  string `json:"result"`
		Errors  Errors `json:"errors"`
	}{}
	err = decoder.Decode(&data)
	if err != nil {
		return "", err
	}
	if len(data.Errors) > 0 {
		return "", data.Errors[0]
	}

	return data.BaseUrl, nil
}
