package emotes

func convertToEmoteSlice[EmoteType Emote](s []EmoteType) []Emote {
	asEmote := make([]Emote, len(s), len(s))
	for i := range s {
		asEmote[i] = Emote(s[i])
	}
	return asEmote
}

type Provider interface {
	IdentifierCode() rune
	LoadGlobalEmotes() ([]Emote, error)
	LoadChannelEmotes(channelID string) ([]Emote, error)
	LoadSpecificEmote(emoteID string) (Emote, error)
}

// Interface type constraints
var _ Provider = &BttvProvider{}
var _ Provider = &FfzProvider{}
var _ Provider = &SevenTVProvider{}

type BttvProvider struct{}

func (b BttvProvider) IdentifierCode() rune {
	return 'b'
}

func (b BttvProvider) LoadGlobalEmotes() ([]Emote, error) {
	res, err := GetGlobalBTTVEmotes()
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil
}

func (b BttvProvider) LoadChannelEmotes(channelID string) ([]Emote, error) {
	res, err := GetChannelBTTVEmotes(channelID)
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil
}

func (b BttvProvider) LoadSpecificEmote(emoteID string) (Emote, error) {
	return GetSpecificBTTVEmote(emoteID)
}

type FfzProvider struct{}

func (f FfzProvider) IdentifierCode() rune {
	return 'f'
}

func (f FfzProvider) LoadGlobalEmotes() ([]Emote, error) {
	res, err := GetGlobalFFZEmotes()
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil
}

func (f FfzProvider) LoadChannelEmotes(channelID string) ([]Emote, error) {
	res, err := GetChannelFFZEmotes(channelID)
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil
}

func (f FfzProvider) LoadSpecificEmote(emoteID string) (Emote, error) {
	return GetSpecificFFZEmote(emoteID)
}

type SevenTVProvider struct{}

func (s SevenTVProvider) IdentifierCode() rune {
	return 's'
}

func (s SevenTVProvider) LoadGlobalEmotes() ([]Emote, error) {
	res, err := GetGlobalSevenTVEmotes()
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil

}

func (s SevenTVProvider) LoadChannelEmotes(channelID string) ([]Emote, error) {
	res, err := GetChannelSevenTVEmotes(channelID)
	if err != nil {
		return nil, err
	}

	return convertToEmoteSlice(res), nil
}

func (s SevenTVProvider) LoadSpecificEmote(emoteID string) (Emote, error) {
	return GetSpecificSevenTVEmote(emoteID)
}
