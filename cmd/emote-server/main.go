package main

import (
	"context"
	"flag"
	"github.com/dnsge/twitch-mobile-emotes"
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
	wsHost := flag.String("ws-host", "irc-ws.chat.twitch.tv", "Host header to expect from Websocket IRC requests")
	emHost := flag.String("emoticon-host", "static-cdn.jtvnw.net", "Host header to expect from Emoticon requests")
	excludeGifs := flag.Bool("no-gifs", false, "Disable showing gif emotes")
	cachePath := flag.String("cache", "", "Path to cache files (leave empty to disable)")
	highRes := flag.Bool("high-res", true, "Whether to always use high-resolution emotes")
	purge := flag.Bool("purge", false, "Purge cache on startup")
	flag.Parse()

	ctx := signalInterrupterContext()
	server := tme.MakeServer(&tme.ServerConfig{
		Address:       *addr,
		WebsocketHost: *wsHost,
		EmoticonHost:  *emHost,
		ExcludeGifs:   *excludeGifs,
		CachePath:     *cachePath,
		HighRes:       *highRes,
		Purge:         *purge,
		Context:       ctx,
	})

	<-ctx.Done()
	timeout, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := server.Shutdown(timeout); err != nil {
		log.Printf("Shutdown error: %v\n", err)
	}
}
