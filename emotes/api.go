package emotes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	bttvGlobalEmotesEndpoint  = "https://api.betterttv.net/3/cached/emotes/global"
	bttvChannelEmotesEndpoint = "https://api.betterttv.net/3/cached/users/twitch/%s"
	ffzGlobalEmotesEndpoint   = "https://api.betterttv.net/3/cached/frankerfacez/emotes/global"
	ffzChannelEmotesEndpoint  = "https://api.betterttv.net/3/cached/frankerfacez/users/twitch/%s"

	userAgent = "twitch-mobile-emotes/0.1"
)

var client = http.Client{
	Timeout: time.Second * 20,
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

func GetGlobalBTTVEmotes() ([]*BttvEmote, error) {
	req, err := http.NewRequest("GET", bttvGlobalEmotesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data []*BttvEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func GetGlobalFFZEmotes() ([]*FfzEmote, error) {
	req, err := http.NewRequest("GET", ffzGlobalEmotesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data []*FfzEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func GetChannelBTTVEmotes(channelID string) ([]*BttvEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(bttvChannelEmotesEndpoint, channelID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data BttvChannelResponse
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	var es []*BttvEmote
	es = append(es, data.ChanEmotes...)
	es = append(es, data.SharedEmotes...)

	return es, nil
}

func GetChannelFFZEmotes(channelID string) ([]*FfzEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(ffzChannelEmotesEndpoint, channelID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data []*FfzEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}
