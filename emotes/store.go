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
	globalBttv   []*BttvEmote
	globalFfz    []*FfzEmote
	channels     map[string]*ChannelEmotes
	channelTimes map[string]time.Time
	wordMaps     map[string]WordMap

	mu sync.Mutex
}

func NewEmoteStore() *EmoteStore {
	return &EmoteStore{
		globalBttv:   nil,
		globalFfz:    nil,
		channels:     make(map[string]*ChannelEmotes),
		channelTimes: make(map[string]time.Time),
		wordMaps:     make(map[string]WordMap),
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

func (s *EmoteStore) FindBttvEmote(emoteID string) (*BttvEmote, bool) {
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

	return nil, false
}

func (s *EmoteStore) FindFfzEmote(emoteID string) (*FfzEmote, bool) {
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

	return nil, false
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
		wordMap[e.Code] = e
	}
	for _, e := range s.globalFfz {
		wordMap[e.Code] = e
	}

	channelEmotes, ok := s.channels[channelID]
	if ok {
		for _, e := range channelEmotes.bttv {
			wordMap[e.Code] = e
		}
		for _, e := range channelEmotes.ffz {
			wordMap[e.Code] = e
		}
	}

	s.wordMaps[channelID] = wordMap
}

func (s *EmoteStore) GetEmoteFromWord(word, channelID string) (Emote, bool) {
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
