package emotes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	bttvGlobalEmotesEndpoint  = "https://api.betterttv.net/3/cached/emotes/global"
	bttvChannelEmotesEndpoint = "https://api.betterttv.net/3/cached/users/twitch/%s"
	bttvSpecificEmoteEndpoint = "https://api.betterttv.net/3/emotes/%s"
	ffzGlobalEmotesEndpoint   = "https://api.frankerfacez.com/v1/set/global"
	ffzChannelEmotesEndpoint  = "https://api.frankerfacez.com/v1/room/id/%s"
	ffzSpecificEmoteEndpoint  = "https://api.frankerfacez.com/v1/emote/%s"

	userAgent = "twitch-mobile-emotes/1.0"
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

func GetSpecificBTTVEmote(emoteID string) (*BttvEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(bttvSpecificEmoteEndpoint, emoteID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data BttvEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return &data, nil
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

	var data FfzGlobal
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	var emotes []*FfzEmote
	for _, setID := range data.DefaultSets {
		setIDAsString := strconv.Itoa(setID)
		set, ok := data.Sets[setIDAsString]
		if !ok {
			return nil, fmt.Errorf("FFZ returned default set of ID %q but didn't provide set", setIDAsString)
		}
		emotes = append(emotes, set.Emoticons...)
	}

	return emotes, nil
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

	var data FfzRoom
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	setIDAsString := strconv.Itoa(data.RoomInfo.Set)
	set, ok := data.Sets[setIDAsString]
	if !ok {
		return nil, fmt.Errorf("FFZ returned room set of ID %q for room %q but didn't provide set", setIDAsString, data.RoomInfo.Set)
	}

	return set.Emoticons, nil
}

func GetSpecificFFZEmote(emoteID string) (*FfzEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(ffzSpecificEmoteEndpoint, emoteID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data FfzEmoteContainer
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return &data.Emote, nil
}
