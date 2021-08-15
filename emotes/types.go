package emotes

import "strconv"

type BttvEmote struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	ImageType string `json:"imageType"`
}

type BttvChannelResponse struct {
	ChanEmotes   []*BttvEmote `json:"channelEmotes"`
	SharedEmotes []*BttvEmote `json:"sharedEmotes"`
}

func (b *BttvEmote) EmoteID() string {
	return b.ID
}

func (b *BttvEmote) LetterCode() string {
	return "b"
}

func (b *BttvEmote) Type() string {
	return b.ImageType
}

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

func (f *FfzEmote) EmoteID() string {
	return strconv.Itoa(f.ID)
}

func (f *FfzEmote) LetterCode() string {
	return "f"
}

func (f *FfzEmote) Type() string {
	return "png" // FFZ only supports pngs at the moment
}

type Emote interface {
	EmoteID() string
	URL(size ImageSize) string
	LetterCode() string
	Type() string
}
