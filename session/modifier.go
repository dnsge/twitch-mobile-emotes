package session

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"strings"
	"unicode/utf8"
)

const leftPrefix = "vl"
const rightPrefix = "vr"

func injectThirdPartyEmotes(emoteStore *emotes.EmoteStore, imageCache *emotes.ImageFileCache, msg *irc.Message, channelID string, includeGifs bool) error {
	messageBody := msg.Trailing()
	emoteTag, err := irc.NewEmoteTag(msg.Tags["emotes"])
	if err != nil {
		return err
	}

	i := 0
	for _, word := range strings.Split(messageBody, " ") {
		wordLen := utf8.RuneCountInString(word) // UTF-8 so emojis don't mess up
		if e, found := emoteStore.GetEmoteFromWord(word, channelID); found {
			if includeGifs || e.Type() != "gif" {
				ratio, err := imageCache.GetEmoteAspectRatio(e)
				if err != nil {
					return err
				}

				if isWide(ratio) && wordLen >= 3 {
					emoteTag.Add(leftPrefix+e.LetterCode()+e.EmoteID(), [2]int{i, i + 1})
					emoteTag.Add(rightPrefix+e.LetterCode()+e.EmoteID(), [2]int{i + 2, i + wordLen - 1})
				} else {
					emoteTag.Add(e.LetterCode()+e.EmoteID(), [2]int{i, i + wordLen - 1})
				}
			}
		}
		i += wordLen + 1
	}

	msg.Tags["emotes"] = emoteTag.TagValue()
	return nil
}

func isWide(ratio float64) bool {
	return ratio >= 1.75
}
