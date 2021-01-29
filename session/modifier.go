package session

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"log"
	"strings"
	"unicode/utf8"
)

const leftPrefix = "vl"
const rightPrefix = "vr"

const commandRune = 0x01

func trimCommandRune(r rune) bool {
	return r == commandRune
}

func injectThirdPartyEmotes(emoteStore *emotes.EmoteStore, imageCache *emotes.ImageFileCache, msg *irc.Message, channelID string, includeGifs bool) error {
	messageBody := msg.Trailing()
	emoteTag, err := irc.NewEmoteTag(msg.Tags["emotes"])
	if err != nil {
		return err
	}

	if messageBody[0] == commandRune {
		messageBody = strings.TrimFunc(messageBody, trimCommandRune)
		spaceIndex := strings.IndexRune(messageBody, ' ')
		if spaceIndex != -1 && spaceIndex < len(messageBody) {
			messageBody = messageBody[spaceIndex + 1:]
		}
	}

	i := 0
	for _, word := range strings.Split(messageBody, " ") {
		wordLen := utf8.RuneCountInString(word) // UTF-8 so emojis don't mess up
		if e, found := emoteStore.GetEmoteFromWord(word, channelID); found {
			if includeGifs || e.Type() != "gif" {
				wide := false
				if imageCache != nil {
					ratio, err := imageCache.GetEmoteAspectRatio(e)
					if err != nil {
						return err
					}
					wide = isWide(ratio)
				}

				if wide && wordLen >= 3 {
					emoteTag.Add(leftPrefix+e.LetterCode()+e.EmoteID(), [2]int{i, i + 1})
					emoteTag.Add(rightPrefix+e.LetterCode()+e.EmoteID(), [2]int{i + 2, i + wordLen - 1})
					go func() {
						err := imageCache.DownloadVirtualToCache(e, emotes.ImageSizeLarge)
						if err != nil {
							log.Printf("Pre-fetch virtual emote: %v\n", err)
						}
					}()
				} else {
					emoteTag.Add(e.LetterCode()+e.EmoteID(), [2]int{i, i + wordLen - 1})
					go func() {
						err := imageCache.DownloadToCache(e, emotes.ImageSizeLarge)
						if err != nil {
							log.Printf("Pre-fetch emote: %v\n", err)
						}
					}()
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
