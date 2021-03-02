package app

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/storage"
)

type Context struct {
	EmoteStore         *emotes.EmoteStore
	ImageCache         *emotes.ImageFileCache
	Config             *ServerConfig
	SettingsRepository storage.SettingsRepository
}
