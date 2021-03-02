package app

import "context"

type ServerConfig struct {
	Address        string
	WebsocketHost  string
	EmoticonHost   string
	IncludeGifs    bool
	CachePath      string
	HighRes        bool
	Purge          bool
	RedisConn      string
	RedisNamespace string
	Context        context.Context
}
