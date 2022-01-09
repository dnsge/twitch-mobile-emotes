package session

import (
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"strings"
)

type Command interface {
	Name() string
	Run(s *wsSession, msg *irc.Message, args []string)
}

type sCommand struct {
	name string
	f    func(s *wsSession, msg *irc.Message, args []string)
}

func SimpleCommand(name string, f func(s *wsSession, msg *irc.Message, args []string)) Command {
	return &sCommand{
		name: name,
		f:    f,
	}
}

func (c *sCommand) Name() string {
	return c.name
}

func (c *sCommand) Run(s *wsSession, msg *irc.Message, args []string) {
	c.f(s, msg, args)
}

func IsCommand(msg *irc.Message, cmd Command, prefix string) bool {
	return msg.Trailing() == prefix+cmd.Name() || strings.HasPrefix(msg.Trailing(), prefix+cmd.Name()+" ")
}

func ExtractCommandArgs(msg *irc.Message, cmd Command, prefix string) []string {
	remaining := msg.Trailing()[len(prefix+cmd.Name()):]
	if len(remaining) == 0 {
		return []string{}
	}

	return strings.Split(remaining[1:], " ")
}
