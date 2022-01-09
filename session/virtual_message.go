package session

import (
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"github.com/google/uuid"
	"strconv"
	"strings"
	"time"
)

type VirtualMessageUser struct {
	DisplayName string
	UserName    string
	UserColor   string
	Badges      []string
}

func buildVirtualMessage(vmUser *VirtualMessageUser, channelName, body string) *irc.Message {
	return &irc.Message{
		Tags: map[string]irc.TagValue{
			"id":           irc.TagValue(uuid.NewString()),
			"user-id":      "1",
			"display-name": irc.TagValue(vmUser.DisplayName),
			"color":        irc.TagValue(vmUser.UserColor),
			"badges":       irc.TagValue(strings.Join(vmUser.Badges, ",")),
			"tmi-sent-ts":  irc.TagValue(strconv.FormatInt(time.Now().UnixMilli(), 10)),
		},
		Prefix:  &irc.Prefix{
			Name: vmUser.UserName,
			User: vmUser.UserName,
			Host: vmUser.UserName + ".tmi.twitch.tv",
		},
		Command: "PRIVMSG",
		Params:  []string{channelName, body},
	}
}
