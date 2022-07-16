package emotes

import (
	"fmt"
	"net/http"
)

const (
	sevenTVGlobalEmotesEndpoint   = "https://api.7tv.app/v2/emotes/global"
	sevenTVChannelEmotesEndpoint  = "https://api.7tv.app/v2/users/%s/emotes"
	sevenTVSpecificEmotesEndpoint = "https://api.7tv.app/v2/emotes/%s"
)

type SevenTVEmote struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	MimeType string      `json:"mime"`
	URLs     [][2]string `json:"urls"`

	Widths  []int `json:"width"`
	Heights []int `json:"height"`
}

func (s *SevenTVEmote) EmoteID() string {
	return s.ID
}

func (s *SevenTVEmote) TypedName() string {
	return s.Name
}

func (s *SevenTVEmote) LetterCode() string {
	return "s"
}

func (s *SevenTVEmote) Type() string {
	return s.MimeType
}

var _ Emote = &SevenTVEmote{}

func GetGlobalSevenTVEmotes() ([]*SevenTVEmote, error) {
	req, err := http.NewRequest("GET", sevenTVGlobalEmotesEndpoint, nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data []*SevenTVEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func GetChannelSevenTVEmotes(channelID string) ([]*SevenTVEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(sevenTVChannelEmotesEndpoint, channelID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound { // SevenTV returns 404 if no channel-specific emotes are set
		_ = resp.Body.Close()
		return []*SevenTVEmote{}, nil
	}

	var data []*SevenTVEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func GetSpecificSevenTVEmote(emoteID string) (*SevenTVEmote, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(sevenTVSpecificEmotesEndpoint, emoteID), nil)
	if err != nil {
		return nil, err
	}

	populateHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var data SevenTVEmote
	if err := unmarshalResponseBody(resp, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
