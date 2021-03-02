package session

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
)

const leftPrefix = "vl"
const rightPrefix = "vr"

const commandRune = 0x01
const CacheDestroyerSize = 8

func init() {
	rand.Seed(time.Now().UnixNano())
}

func trimCommandRune(r rune) bool {
	return r == commandRune
}

func injectThirdPartyEmotes(s *wsSession, msg *irc.Message, channelID string) error {
	messageBody := msg.Trailing()
	emoteTag, err := irc.NewEmoteTag(msg.Tags["emotes"])
	if err != nil {
		return err
	}

	if messageBody[0] == commandRune {
		messageBody = strings.TrimFunc(messageBody, trimCommandRune)
		spaceIndex := strings.IndexRune(messageBody, ' ')
		if spaceIndex != -1 && spaceIndex < len(messageBody) {
			messageBody = messageBody[spaceIndex+1:]
		}
	}

	i := 0
	for _, word := range strings.Split(messageBody, " ") {
		wordLen := utf8.RuneCountInString(word) // UTF-8 so emojis don't mess up
		if e, found := s.emoteStore.GetEmoteFromWord(word, channelID); found {
			if s.showGifs() || e.Type() != "gif" {
				wide := false
				if s.imageCache != nil {
					ratio, err := s.imageCache.GetEmoteAspectRatio(e)
					if err != nil {
						return err
					}
					wide = isWide(ratio)
				}

				cacheDestroyerPrefix := ""
				if s.settings.CacheDestroyerKey != "" {
					if len(s.settings.CacheDestroyerKey) != CacheDestroyerSize {
						s.settings.CacheDestroyerKey = newCacheDestroyer(CacheDestroyerSize)
					}
					cacheDestroyerPrefix = "d" + s.settings.CacheDestroyerKey
				}

				if wide && wordLen >= 3 {
					emoteTag.Add(cacheDestroyerPrefix+leftPrefix+e.LetterCode()+e.EmoteID(), [2]int{i, i + 1})
					emoteTag.Add(cacheDestroyerPrefix+rightPrefix+e.LetterCode()+e.EmoteID(), [2]int{i + 2, i + wordLen - 1})
					go func() {
						err := s.imageCache.DownloadVirtualToCache(e, emotes.ImageSizeLarge)
						if err != nil {
							log.Printf("Pre-fetch virtual emote: %v\n", err)
						}
					}()
				} else {
					emoteTag.Add(cacheDestroyerPrefix+e.LetterCode()+e.EmoteID(), [2]int{i, i + wordLen - 1})
					if s.imageCache != nil {
						go func() {
							err := s.imageCache.DownloadToCache(e, emotes.ImageSizeLarge)
							if err != nil {
								log.Printf("Pre-fetch emote: %v\n", err)
							}
						}()
					}
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
func newCacheDestroyer(size int) string {
	b := make([]rune, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
