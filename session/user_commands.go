package session

import (
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"log"
	"strings"
)

const commandPrefix = "@@"

var helpText = "Twitch Mobile Emotes Help\n" +
	"@@reload - Reload BTTV and FFZ emotes\n" +
	"@@cache - Set a cache destroyer value (warning: unstable)\n" +
	"@@cache off - Disable cache destroyer\n" +
	"@@gifs [on|off] - Enable or disable GIF emotes\n" +
	"@@help - Show this message"

var systemUser = &VirtualMessageUser{
	DisplayName: "Mobile Emotes",
	UserName:    "mobile_emotes",
	UserColor:   "#FF0000",
	Badges:      []string{"staff/1", "broadcaster/1", "moderator/1"},
}

var allCommands = []Command{
	SimpleCommand("reload", func(s *wsSession, msg *irc.Message, args []string) {
		channelName := strings.ToLower(msg.Params[0])
		channelID, found := channelNameMap[channelName]
		if found {
			err := s.emoteStore.Load(channelID)
			if err != nil {
				log.Printf("Error reloading channel: %v\n", err)
				return
			} else {
				var body string
				if s.state.Greeted {
					body = "@" + s.state.Username + ", reloaded BTTV and FFZ emotes. The old emote images may remain cached on your device."
				} else { // really shouldn't be possible
					body = "Reloaded BTTV and FFZ emotes. The old emote images may remain cached on your device."
				}

				// try to catch the eye with fancy badges
				s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], body))
			}
		}
	}),
	SimpleCommand("cache", func(s *wsSession, msg *irc.Message, args []string) {
		if s.settings == nil {
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Error: User settings are not enabled"))
			return
		}

		if len(args) == 1 && args[0] == "off" {
			s.settings.CacheDestroyerKey = ""
			s.saveSettings()
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Removed cache destroyer value"))
			return
		}

		s.settings.CacheDestroyerKey = newCacheDestroyer(CacheDestroyerSize)
		var body string
		if s.state.Greeted {
			body = "@" + s.state.Username + ", set new cache destroyer value to " + s.settings.CacheDestroyerKey
		} else {
			body = "Set new cache destroyer value to " + s.settings.CacheDestroyerKey
		}

		s.saveSettings()
		s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], body))
	}),
	SimpleCommand("gifs", func(s *wsSession, msg *irc.Message, args []string) {
		if s.settings == nil {
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Error: User settings are not enabled"))
			return
		}

		if len(args) != 1 || (args[0] != "on" && args[0] != "off") {
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Usage: gifs [on|off]"))
			return
		}

		if args[0] == "on" {
			s.settings.EnableGifEmotes = true
			s.saveSettings()
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Enabled gif emotes"))
		} else if args[0] == "off" {
			s.settings.EnableGifEmotes = false
			s.saveSettings()
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], "Disabled gif emotes"))
		}
	}),
	SimpleCommand("help", func(s *wsSession, msg *irc.Message, args []string) {
		lines := strings.Split(helpText, "\n")
		for i := range lines {
			s.writeClientMessage(1, buildVirtualMessage(systemUser, msg.Params[0], lines[i]))
		}
	}),
}
