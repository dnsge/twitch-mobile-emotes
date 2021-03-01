package emotes

import (
	"strconv"
	"sync"
	"time"
)

const (
	cachedEmoteDuration = time.Hour
)

type ChannelEmotes struct {
	bttv []*BttvEmote
	ffz  []*FfzEmote
}

type WordMap map[string]Emote

type EmoteStore struct {
	// Globally available emotes
	globalBttv     []*BttvEmote
	globalFfz      []*FfzEmote

	// Emotes that were requested but not found in any other channel
	danglingEmotes *ChannelEmotes

	// Emotes belonging to channels
	channels       map[string]*ChannelEmotes
	channelTimes   map[string]time.Time
	wordMaps       map[string]WordMap

	mu sync.Mutex
}

func NewEmoteStore() *EmoteStore {
	return &EmoteStore{
		globalBttv:     nil,
		globalFfz:      nil,
		danglingEmotes: &ChannelEmotes{},
		channels:       make(map[string]*ChannelEmotes),
		channelTimes:   make(map[string]time.Time),
		wordMaps:       make(map[string]WordMap),
	}
}

func (s *EmoteStore) Init() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	globalB, err := GetGlobalBTTVEmotes()
	if err != nil {
		return err
	}
	s.globalBttv = globalB

	globalF, err := GetGlobalFFZEmotes()
	if err != nil {
		return err
	}
	s.globalFfz = globalF

	return nil
}

// Loads and caches BTTV and FFZ emotes for a channel
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
	bttv, err := GetChannelBTTVEmotes(channelID)
	if err != nil {
		return err
	}
	ffz, err := GetChannelFFZEmotes(channelID)
	if err != nil {
		return err
	}

	s.channels[channelID] = &ChannelEmotes{
		bttv: bttv,
		ffz:  ffz,
	}
	s.channelTimes[channelID] = time.Now()
	s.updateWordMap(channelID)
	return nil
}

func (s *EmoteStore) GetChannelEmotes(channelID string) (*ChannelEmotes, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.channels[channelID]
	return e, ok
}

func (s *EmoteStore) GetBttvEmote(emoteID string) (*BttvEmote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.globalBttv {
		if e.ID == emoteID {
			return e, true
		}
	}

	for _, c := range s.channels {
		for _, e := range c.bttv {
			if e.ID == emoteID {
				return e, true
			}
		}
	}

	for _, e := range s.danglingEmotes.bttv {
		if e.ID == emoteID {
			return e, true
		}
	}

	e, err := GetSpecificBTTVEmote(emoteID)
	if err != nil {
		return nil, false
	} else {
		s.danglingEmotes.bttv = append(s.danglingEmotes.bttv, e)
		return e, true
	}
}

func (s *EmoteStore) GetFfzEmote(emoteID string) (*FfzEmote, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	intID, err := strconv.Atoi(emoteID)
	if err != nil {
		return nil, false
	}

	for _, e := range s.globalFfz {
		if e.ID == intID {
			return e, true
		}
	}

	for _, c := range s.channels {
		for _, e := range c.ffz {
			if e.ID == intID {
				return e, true
			}
		}
	}

	for _, e := range s.danglingEmotes.ffz {
		if e.ID == intID {
			return e, true
		}
	}

	e, err := GetSpecificFFZEmote(emoteID)
	if err != nil {
		return nil, false
	} else {
		s.danglingEmotes.ffz = append(s.danglingEmotes.ffz, e)
		return e, true
	}
}

func (s *EmoteStore) updateWordMap(channelID string) {
	// priority:
	// FFZ Channel
	// BTTV Channel
	// FFZ Global
	// BTTV Global
	// Work in reverse order so later ones override earlier ones

	wordMap := make(WordMap)
	for _, e := range s.globalBttv {
		if old, found := wordMap[e.Code]; found { // prioritize gif emotes of same name
			if old.Type() != "gif" {
				wordMap[e.Code] = e
			}
		} else {
			wordMap[e.Code] = e
		}
	}
	for _, e := range s.globalFfz {
		wordMap[e.Name] = e
	}

	channelEmotes, ok := s.channels[channelID]
	if ok {
		for _, e := range channelEmotes.bttv {
			if old, found := wordMap[e.Code]; found { // prioritize gif emotes of same name
				if old.Type() != "gif" {
					wordMap[e.Code] = e
				}
			} else {
				wordMap[e.Code] = e
			}
		}
		for _, e := range channelEmotes.ffz {
			wordMap[e.Name] = e
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
