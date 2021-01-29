package app

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
)

type Context struct {
	EmoteStore *emotes.EmoteStore
	ImageCache *emotes.ImageFileCache
	Config     *ServerConfig
}
