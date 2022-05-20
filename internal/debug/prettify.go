package debug

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
)

func prettifyJSON(text []byte) string {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, text, "", "    ")
	if err != nil {
		panic(err)
	}
	return prettyJSON.String()
}

func PrettifyJSONFromReader(r io.Reader) string {
	bd, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return prettifyJSON(bd)
}

func PrettifyJSONFromStruct(v any) string {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
