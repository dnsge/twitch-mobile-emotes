package tme

import (
	"github.com/dnsge/twitch-mobile-emotes/app"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/storage"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"time"
)

func MakeServer(cfg *app.ServerConfig) *http.Server {
	s := &http.Server{
		Addr:    cfg.Address,
		Handler: handleRequest(cfg),
	}

	go func() {
		log.Printf("Starting server on %s\n", cfg.Address)
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	return s
}

func handleRequest(cfg *app.ServerConfig) http.HandlerFunc {
	store := emotes.NewEmoteStore()
	if err := store.Init(); err != nil {
		log.Fatalln(err)
	}

	var cache *emotes.ImageFileCache = nil
	if cfg.CachePath != "" { // cache is enabled
		cache = emotes.NewImageFileCache(cfg.CachePath, time.Hour*48, true)
		if err := cache.Index(); err != nil {
			log.Fatalln(err)
		}

		if cfg.Purge {
			if err := cache.Purge(); err != nil {
				log.Fatalf("Purge cache: %v\n", err)
			}
		}

		go cache.AutoEvict(cfg.Context)
	}

	var settingsRepository storage.SettingsRepository = nil
	if cfg.RedisConn != "" {
		opts, err := redis.ParseURL(cfg.RedisConn)
		if err != nil {
			log.Fatalf("Parse redis URL: %v\n", err)
		}
		settingsRepository = storage.NewRedisSettingsRepository(cfg.RedisNamespace, opts, cfg.Context)
		log.Println("Connected to Redis")
	}

	appCtx := &app.Context{
		EmoteStore:         store,
		ImageCache:         cache,
		Config:             cfg,
		SettingsRepository: settingsRepository,
	}

	manager := NewWsForwarder(appCtx)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == cfg.WebsocketHost {
			manager.HandleWsConnection(w, r)
		} else if r.Host == cfg.EmoticonHost {
			handleEmoticonRequest(w, r, store, cache)
		} else {
			log.Printf("Got unexpected Host value %q\n", r.Host)
			http.NotFound(w, r)
		}
	}
}
