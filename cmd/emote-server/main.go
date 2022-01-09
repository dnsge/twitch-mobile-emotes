package main

import (
	"context"
	"flag"
	"github.com/dnsge/twitch-mobile-emotes"
	"github.com/dnsge/twitch-mobile-emotes/app"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func signalInterrupterContext() context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer func() {
			cancel()
		}()
		<-c
	}()

	return ctx
}

func main() {
	addr := flag.String("address", "0.0.0.0:8080", "Bind address")
	debug := flag.Bool("debug", false, "Enable debug printouts")
	wsHost := flag.String("ws-host", "irc-ws.chat.twitch.tv", "Host header to expect from Websocket IRC requests")
	emHost := flag.String("emoticon-host", "static-cdn.jtvnw.net", "Host header to expect from Emoticon requests")
	excludeGifs := flag.Bool("no-gifs", false, "Disable showing gif emotes")
	cachePath := flag.String("cache", "", "Path to cache files (leave empty to disable)")
	purge := flag.Bool("purge", false, "Purge cache on startup")
	idealGifsFile := flag.String("ideal-gifs", "", "Path to ideal gif frames file (leave empty to disable)")
	redisConn := flag.String("redis-url", "", "Redis connection string")
	redisNamespace := flag.String("redis-namespace", "tme", "Redis key namespace")
	flag.Parse()

	if *idealGifsFile != "" {
		emotes.InitIdealGifFrames(*idealGifsFile)
	}

	ctx := signalInterrupterContext()
	server := tme.MakeServer(&app.ServerConfig{
		Address:        *addr,
		Debug:          *debug,
		WebsocketHost:  *wsHost,
		EmoticonHost:   *emHost,
		IncludeGifs:    !*excludeGifs,
		CachePath:      *cachePath,
		Purge:          *purge,
		RedisConn:      *redisConn,
		RedisNamespace: *redisNamespace,
		Context:        ctx,
	})

	<-ctx.Done()
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Shutdown(timeout); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	}
}
