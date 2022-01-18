package main

import (
	"context"
	"flag"
	"fmt"
	tme "github.com/dnsge/twitch-mobile-emotes"
	"github.com/dnsge/twitch-mobile-emotes/app"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/session"
	"github.com/dnsge/twitch-mobile-emotes/storage"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	auth           = flag.String("auth", "", "AUTH to pass to Twitch IRC (format of oauth:xxx)")
	nick           = flag.String("nick", "", "NICK to pass to Twitch IRC (Twitch username)")
	requestCaps    = flag.Bool("req-caps", true, "Request Twitch capabilities")
	redisConn      = flag.String("redis-url", "", "Redis connection string")
	redisNamespace = flag.String("redis-namespace", "tme", "Redis key namespace")
)

func init() {
	flag.Parse()
}

func main() {
	if *auth == "" {
		fmt.Printf("Error: --auth flag is required")
		os.Exit(1)
	}

	if *nick == "" {
		fmt.Printf("Error: --nick flag is required")
		os.Exit(1)
	}

	ctx := signalInterrupterContext()

	twitchConn, err := tme.ConnectToTwitchIrc(ctx)
	if err != nil {
		log.Printf("Failed to connect to Twitch IRC server: %v\n", err)
		os.Exit(1)
		return
	}

	var settingsRepository storage.SettingsRepository = nil
	if *redisConn != "" {
		opts, err := redis.ParseURL(*redisConn)
		if err != nil {
			log.Fatalf("Parse redis URL: %v\n", err)
		}
		r := storage.NewRedisSettingsRepository(*redisNamespace, opts, ctx)
		if err := r.Ping(); err != nil {
			log.Fatalf("Failed to communicate with Redis: %v\n", err)
		}
		settingsRepository = r
		log.Println("Connected to Redis")
	}

	store := emotes.NewEmoteStore()
	if err := store.Init(); err != nil {
		log.Fatalln(err)
	}

	appCtx := &app.Context{
		EmoteStore: store,
		ImageCache: nil,
		Config: &app.ServerConfig{
			Debug:       false,
			IncludeGifs: true,
			Context:     ctx,
		},
		SettingsRepository: settingsRepository,
	}

	consoleConn := NewConsoleConn()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		session.RunWsSession(consoleConn, twitchConn, appCtx)
		wg.Done()
	}()

	if err := twitchConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("PASS %s\r\n", *auth))); err != nil {
		log.Printf("Failed to send AUTH message to Twitch IRC server: %v\n", err)
		os.Exit(1)
	}

	if err := twitchConn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("NICK %s\r\n", *nick))); err != nil {
		log.Printf("Failed to send PASS message to Twitch IRC server: %v\n", err)
		os.Exit(1)
	}

	if *requestCaps {
		if err := twitchConn.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/tags twitch.tv/commands\r\n")); err != nil {
			log.Printf("Failed to request CAPS from Twitch IRC server: %v\n", err)
			os.Exit(1)
		}
	}

	wg.Wait()
}

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
