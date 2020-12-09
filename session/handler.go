package session

import (
	"fmt"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"log"
)

// returns whether the message was passed on and an error
func (s *wsSession) handleTwitchMessage(msg *irc.Message) (bool, error) {
	if msg.Command == "PRIVMSG" || msg.Command == "USERNOTICE" {
		channelID, found := msg.GetTag("room-id")
		if !found {
			return false, fmt.Errorf("missing user id tag for %q", msg.Params[0])
		}

		if err := injectThirdPartyEmotes(s.emoteStore, msg, channelID, s.includeGifs); err != nil {
			return false, fmt.Errorf("inject emotes: %w", err)
		}

		return true, nil // modified
	} else if msg.Command == "ROOMSTATE" {
		channelID, found := msg.GetTag("room-id")
		if !found {
			return false, fmt.Errorf("missing user id tag for %q", msg.Params[0])
		}

		if err := s.emoteStore.LoadIfNotLoaded(channelID); err != nil {
			return false, fmt.Errorf("load channel: %w", err)
		}
	}

	return false, nil
}

// returns whether the message should be passed on, whether it was modified, and an error
func (s *wsSession) handleClientMessage(msg *irc.Message) (bool, bool, error) {
	if !s.greeted && msg.Command == "NICK" {
		userName := msg.Params[0]
		log.Printf("User %q connected\n", userName)
		s.greeted = true
	}

	return true, false, nil
}
