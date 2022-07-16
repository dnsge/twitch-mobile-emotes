package emotes

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	ffzGlobalEmotesEndpoint  = "https://api.frankerfacez.com/v1/set/global"
	ffzChannelEmotesEndpoint = "https://api.frankerfacez.com/v1/room/id/%s"
	ffzSpecificEmoteEndpoint = "https://api.frankerfacez.com/v1/emote/%s"
)

type FfzGlobal struct {
	DefaultSets []int              `json:"default_sets"`
	Sets        map[string]*FfzSet `json:"sets"`
}

type FfzRoom struct {
	RoomInfo FfzRoomInfo        `json:"room"`
	Sets     map[string]*FfzSet `json:"sets"`
}

type FfzRoomInfo struct {
	Set int `json:"set"`
}

type FfzSet struct {
	Emoticons []*FfzEmote `json:"emoticons"`
}

type FfzUrls struct {
	One  string `json:"1"`
	Two  string `json:"2"`
	Four string `json:"4"`
}

type FfzEmoteContainer struct {
	Emote FfzEmote `json:"emote"`
}

type FfzEmote struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Images FfzUrls `json:"urls"`
}

var _ Emote = &FfzEmote{}

func (f *FfzEmote) EmoteID() string {
	return strconv.Itoa(f.ID)
}

func (f *FfzEmote) TypedName() string {
	return f.Name
}

func (f *FfzEmote) LetterCode() string {
	return "f"
}

func (f *FfzEmote) Type() string {
	return "png" // FFZ only supports pngs at the moment
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
