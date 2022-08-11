package debug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func prettifyJSON(text []byte) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, text, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(prettyJSON.String())
}

func PrintJSONFromReader(r io.Reader) {
	bd, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	prettifyJSON(bd)
}

func PrintJSONFromStruct(v any) {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func PrintRespBody(r *http.Response) {
	bd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bd))
}
