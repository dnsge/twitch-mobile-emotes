package emotes

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	userAgent = "twitch-mobile-emotes/1.0"
)

var client = http.Client{
	Timeout: time.Second * 5,
}

func populateHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
}

func unmarshalResponseBody(resp *http.Response, data interface{}) error {
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, data)
}
