package app

import "context"

type ServerConfig struct {
	Address        string
	WebsocketHost  string
	EmoticonHost   string
	IncludeGifs    bool
	CachePath      string
	Purge          bool
	RedisConn      string
	RedisNamespace string
	Context        context.Context
}
