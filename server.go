package tme

import (
	"context"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"log"
	"net/http"
	"time"
)

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

func MakeServer(cfg *ServerConfig) *http.Server {
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

func handleRequest(cfg *ServerConfig) http.HandlerFunc {
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

	manager := NewWsForwarder(store, !cfg.ExcludeGifs, cfg.Context)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == cfg.WebsocketHost {
			manager.HandleWsConnection(w, r)
		} else if r.Host == cfg.EmoticonHost {
			handleEmoticonRequest(w, r, store, cache, cfg.HighRes)
		} else {
			log.Printf("Got unexpected Host value %q\n", r.Host)
			http.NotFound(w, r)
		}
	}
}
