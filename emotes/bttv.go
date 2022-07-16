package emotes

import (
	"fmt"
	"net/http"
)

const (
	bttvGlobalEmotesEndpoint  = "https://api.betterttv.net/3/cached/emotes/global"
	bttvChannelEmotesEndpoint = "https://api.betterttv.net/3/cached/users/twitch/%s"
	bttvSpecificEmoteEndpoint = "https://api.betterttv.net/3/emotes/%s"
)

type BttvEmote struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	ImageType string `json:"imageType"`
}

var _ Emote = &BttvEmote{}

type BttvChannelResponse struct {
	ChanEmotes   []*BttvEmote `json:"channelEmotes"`
	SharedEmotes []*BttvEmote `json:"sharedEmotes"`
}

func (b *BttvEmote) EmoteID() string {
	return b.ID
}

func (b *BttvEmote) TypedName() string {
	return b.Code
}

func (b *BttvEmote) LetterCode() string {
	return "b"
}

func (b *BttvEmote) Type() string {
	return b.ImageType
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
