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
		for _, command := range allCommands {
			if IsCommand(msg, command, commandPrefix) {
				args := ExtractCommandArgs(msg, command, commandPrefix)
				command.Run(s, msg, args)
				return false, false, nil
			}
		}
	}

	return true, false, nil
}
