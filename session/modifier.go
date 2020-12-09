package session

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"strings"
	"unicode/utf8"
)

func injectThirdPartyEmotes(emoteStore *emotes.EmoteStore, msg *irc.Message, channelID string, includeGifs bool) error {
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
				emoteTag.Add(e.LetterCode()+e.EmoteID(), [2]int{i, i + wordLen - 1})
			}
		}
		i += wordLen + 1
	}

	msg.Tags["emotes"] = emoteTag.TagValue()
	return nil
}
