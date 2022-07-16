package emotes

import (
	"sync"
	"time"
)

const (
	cachedEmoteDuration = time.Hour
)

type ProviderEmotes map[rune][]Emote
type WordMap map[string]Emote

type EmoteStore struct {
	providers []Provider

	// Globally available emotes
	globalEmotes ProviderEmotes

	// Emotes that were requested but not found in any other channel
	danglingEmotes ProviderEmotes

	// Emotes belonging to channels
	channels     map[string]ProviderEmotes
	channelTimes map[string]time.Time
	wordMaps     map[string]WordMap

	mu sync.Mutex
}

func NewEmoteStore() *EmoteStore {
	return &EmoteStore{
		providers: []Provider{
			&BttvProvider{},
			&FfzProvider{},
			&SevenTVProvider{},
		},
		globalEmotes:   make(ProviderEmotes),
		danglingEmotes: make(ProviderEmotes),
		channels:       make(map[string]ProviderEmotes),
		channelTimes:   make(map[string]time.Time),
		wordMaps:       make(map[string]WordMap),
	}
}

func (s *EmoteStore) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, provider := range s.providers {
		globals, err := provider.LoadGlobalEmotes()
		if err != nil {
			return err
		}
		s.globalEmotes[provider.IdentifierCode()] = globals
	}

	return nil
}

// LoadIfNotLoaded loads and caches BTTV and FFZ emotes for a channel
func (s *EmoteStore) LoadIfNotLoaded(channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.channels[channelID]; ok { // already loaded
		if time.Since(s.channelTimes[channelID]) > cachedEmoteDuration { // emotes are possibly out of date
			delete(s.channels, channelID)
			delete(s.channelTimes, channelID)
		} else {
			return nil
		}
	}
	return s.load(channelID)
}

func (s *EmoteStore) Load(channelID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.load(channelID)
}

func (s *EmoteStore) load(channelID string) error {
	channelEmotes := make(ProviderEmotes)
	for _, provider := range s.providers {
		code := provider.IdentifierCode()
		emotes, err := provider.LoadChannelEmotes(channelID)
		if err != nil {
			return err
		}
		channelEmotes[code] = emotes
	}

	s.channels[channelID] = channelEmotes
	s.channelTimes[channelID] = time.Now()
	s.updateWordMap(channelID)
	return nil
}

func (s *EmoteStore) GetChannelEmotes(channelID string) ([]Emote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.channels[channelID]

	var emotes []Emote
	for _, v := range e {
		emotes = append(emotes, v...)
	}

	return emotes, ok
}

func (s *EmoteStore) ProviderFromCode(identifierCode rune) (Provider, bool) {
	for _, p := range s.providers {
		if p.IdentifierCode() == identifierCode {
			return p, true
		}
	}
	return nil, false
}

func (s *EmoteStore) GetEmote(identifierCode rune, emoteID string) (Emote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if globalList, ok := s.globalEmotes[identifierCode]; !ok {
		return nil, false
	} else {
		// Check the provider's global emotes
		for _, e := range globalList {
			if e.EmoteID() == emoteID {
				return e, true
			}
		}
	}

	for _, channel := range s.channels {
		if channelList, ok := channel[identifierCode]; !ok {
			return nil, false
		} else {
			// Check the provider's channel emotes
			for _, e := range channelList {
				if e.EmoteID() == emoteID {
					return e, true
				}
			}
		}
	}

	provider, ok := s.ProviderFromCode(identifierCode)
	if !ok {
		return nil, false
	}

	e, err := provider.LoadSpecificEmote(emoteID)
	if err != nil {
		return nil, false
	} else {
		s.danglingEmotes[identifierCode] = append(s.danglingEmotes[identifierCode], e)
		return e, true
	}
}

func (s *EmoteStore) updateWordMap(channelID string) {
	// Word map priority is the order of providers.
	// Work in reverse order so later ones override earlier ones!

	wordMap := make(WordMap)

	for i := len(s.providers) - 1; i >= 0; i-- {
		identifierCode := s.providers[i].IdentifierCode()
		emotes, ok := s.globalEmotes[identifierCode]
		if !ok {
			continue
		}

		for _, e := range emotes {
			wordMap[e.TypedName()] = e
		}
	}

	channelEmotes, ok := s.channels[channelID]
	if ok {
		for i := len(s.providers) - 1; i >= 0; i-- {
			identifierCode := s.providers[i].IdentifierCode()
			emotes, ok := channelEmotes[identifierCode]
			if !ok {
				continue
			}

			for _, e := range emotes {
				wordMap[e.TypedName()] = e
			}
		}
	}

	s.wordMaps[channelID] = wordMap
}

func (s *EmoteStore) GetEmoteFromWord(word, channelID string) (Emote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	channelWords, ok := s.wordMaps[channelID]
	if !ok {
		return nil, false
	}

	emote, ok := channelWords[word]
	if !ok {
		return nil, false
	}

	return emote, true
}
