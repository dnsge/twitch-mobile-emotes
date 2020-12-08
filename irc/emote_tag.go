package irc

import (
	"fmt"
	"strconv"
	"strings"
)

type IndexPair [2]int
type IndexList []IndexPair
type EmoteMap = map[string]IndexList

func parseIndexPair(p string) (IndexPair, error) { // form of a-b
	parts := strings.SplitN(p, "-", 2)
	one, err := strconv.Atoi(parts[0])
	if err != nil {
		return [2]int{}, err
	}

	two, err := strconv.Atoi(parts[1])
	if err != nil {
		return [2]int{}, err
	}

	return [2]int{one, two}, nil
}

func (p IndexPair) String() string {
	return fmt.Sprintf("%d-%d", p[0], p[1])
}

func (l IndexList) String() string {
	ret := make([]string, len(l))
	for i, v := range l {
		ret[i] = v.String()
	}
	return strings.Join(ret, ",")
}

type EmoteTag struct {
	Emotes EmoteMap
}

func NewEmoteTag(existing TagValue) (*EmoteTag, error) {
	tag := &EmoteTag{
		Emotes: make(EmoteMap),
	}

	if existing == "" {
		return tag, nil
	}

	for _, emotePart := range strings.Split(string(existing), "/") {
		split := strings.SplitN(emotePart, ":", 2)
		emoteID := split[0]

		pairs := strings.Split(split[1], ",")
		indexes := make(IndexList, len(pairs))

		for i, pair := range pairs {
			parsed, err := parseIndexPair(pair)
			if err != nil {
				return nil, err
			}

			indexes[i] = parsed
		}

		tag.Emotes[emoteID] = indexes
	}

	return tag, nil
}

func (e *EmoteTag) String() string {
	emoteStrings := make([]string, len(e.Emotes))
	n := 0
	for emoteID, emoteIndexes := range e.Emotes {
		emoteStrings[n] = emoteID + ":" + emoteIndexes.String()
		n++
	}
	return strings.Join(emoteStrings, "/")
}

func (e *EmoteTag) TagValue() TagValue {
	return TagValue(e.String())
}

func (e *EmoteTag) Add(emoteID string, index IndexPair) {
	if val, ok := e.Emotes[emoteID]; ok {
		e.Emotes[emoteID] = append(val, index)
	} else {
		e.Emotes[emoteID] = []IndexPair{index}
	}
}
