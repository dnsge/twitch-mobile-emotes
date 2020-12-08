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

type FfzUrls struct {
	One  string `json:"1x"`
	Two  string `json:"2x"`
	Four string `json:"4x"`
}

type FfzEmote struct {
	ID        int     `json:"id"`
	Code      string  `json:"code"`
	Images    FfzUrls `json:"images"`
	ImageType string  `json:"imageType"`
}

func (f *FfzEmote) EmoteID() string {
	return strconv.Itoa(f.ID)
}

func (f *FfzEmote) LetterCode() string {
	return "f"
}

func (f *FfzEmote) Type() string {
	return f.ImageType
}

type Emote interface {
	EmoteID() string
	URL(size ImageSize) string
	LetterCode() string
	Type() string
}
