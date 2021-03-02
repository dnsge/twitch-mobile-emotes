package session

import (
	"fmt"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"github.com/dnsge/twitch-mobile-emotes/storage"
	"log"
	"strings"
)

var channelNameMap = make(map[string]string)

// returns whether the message was passed on and an error
func (s *wsSession) handleTwitchMessage(msg *irc.Message) (bool, error) {
	if msg.Command == "PRIVMSG" || msg.Command == "USERNOTICE" {
		channelID, found := msg.GetTag("room-id")
		if !found {
			return false, fmt.Errorf("missing user id tag for %q", msg.Params[0])
		}

		if err := injectThirdPartyEmotes(s, msg, channelID); err != nil {
			return false, fmt.Errorf("inject emotes: %w", err)
		}

		return true, nil // modified
	} else if msg.Command == "ROOMSTATE" {
		channelID, found := msg.GetTag("room-id")
		if !found {
			return false, fmt.Errorf("missing user id tag for %q", msg.Params[0])
		}

		channelName := strings.ToLower(msg.Params[0])
		channelNameMap[channelName] = channelID

		if err := s.emoteStore.LoadIfNotLoaded(channelID); err != nil {
			return false, fmt.Errorf("load channel: %w", err)
		}
	}

	return false, nil
}

const reloadCommand = "@@reload"
const destroyCacheCommand = "@@cache"

// returns whether the message should be passed on, whether it was modified, and an error
func (s *wsSession) handleClientMessage(msg *irc.Message) (bool, bool, error) {
	switch msg.Command {
	case "PASS":
		go func() {
			userID, err := GetUserIDFromOAuth(msg.Params[0])
			if err != nil {
				log.Printf("Error getting User ID: %v\n", err)
				return
			}

			s.state.UserID = userID
			if s.settingsRepository == nil {
				return
			}

			settings, err := s.settingsRepository.Load(userID)
			if err != nil {
				log.Printf("Error loading settings: %v\n", err)
				return
			}

			if settings == nil { // load default settings
				s.settings = &storage.Settings{
					CacheDestroyerKey: "",
					EnableGifEmotes:   s.defaultIncludeGifs,
				}
				s.saveSettings()
			} else {
				s.settings = settings
			}
		}()
	case "NICK":
		if !s.state.Greeted {
			s.state.Username = msg.Params[0]
			s.state.Greeted = true
		}
	case "PRIVMSG":
		if msg.Trailing() == reloadCommand {
			channelName := strings.ToLower(msg.Params[0])
			channelID, found := channelNameMap[channelName]
			if found {
				err := s.emoteStore.Load(channelID)
				if err != nil {
					return true, false, fmt.Errorf("reload channel: %w", err)
				} else {
					var body string
					if s.state.Greeted {
						body = "@" + s.state.Username + ", reloaded BTTV and FFZ emotes. The old emote images may remain cached on your device."
					} else { // really shouldn't be possible
						body = "Reloaded BTTV and FFZ emotes. The old emote images may remain cached on your device."
					}

					// try to catch the eye with fancy badges
					s.writeClientMessage(1, makeVirtualMessage("staff/1,partner/1,broadcaster/1", msg.Params[0], body))
					return false, false, nil // don't forward the reload message
				}
			}
		} else if strings.HasPrefix(msg.Trailing(), destroyCacheCommand) {
			if msg.Trailing() == destroyCacheCommand+" off" {
				s.settings.CacheDestroyerKey = ""
				s.saveSettings()
				s.writeClientMessage(1, makeVirtualMessage("staff/1,partner/1,broadcaster/1", msg.Params[0], "Removed cache destroyer value"))
				return false, false, nil // don't forward the cache message
			}

			s.settings.CacheDestroyerKey = newCacheDestroyer(CacheDestroyerSize)
			var body string
			if s.state.Greeted {
				body = "@" + s.state.Username + ", set new cache destroyer value to " + s.settings.CacheDestroyerKey
			} else {
				body = "Set new cache destroyer value to " + s.settings.CacheDestroyerKey
			}

			s.saveSettings()
			s.writeClientMessage(1, makeVirtualMessage("staff/1,partner/1,broadcaster/1", msg.Params[0], body))
			return false, false, nil // don't forward the cache message
		} else if msg.Trailing() == "@@debug" {
			fmt.Println(s.state)
			fmt.Println(s.settings)
			return false, false, nil
		}
	}

	return true, false, nil
}

func makeVirtualMessage(badges irc.TagValue, channelName, body string) *irc.Message {
	return &irc.Message{
		Tags:    map[string]irc.TagValue{"badges": badges},
		Command: "PRIVMSG",
		Params:  []string{channelName, body},
	}
}
