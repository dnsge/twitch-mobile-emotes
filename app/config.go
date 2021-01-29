package app

import "context"

type ServerConfig struct {
	Address       string
	WebsocketHost string
	EmoticonHost  string
	ExcludeGifs   bool
	CachePath     string
	HighRes       bool
	Purge         bool
	Context       context.Context
}
