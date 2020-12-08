package tme

import (
	"context"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"log"
	"net/http"
)

type ServerConfig struct {
	Address       string
	WebsocketHost string
	EmoticonHost  string
	ExcludeGifs   bool
	Context       context.Context
}

func MakeServer(cfg *ServerConfig) *http.Server {
	s := &http.Server{
		Addr:    cfg.Address,
		Handler: handleRequest(cfg),
	}

	go func() {
		log.Println("Starting server")
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

	manager := NewWsForwarder(store, !cfg.ExcludeGifs, cfg.Context)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Host == cfg.WebsocketHost {
			manager.HandleWsConnection(w, r)
		} else if r.Host == cfg.EmoticonHost {
			handleEmoticonRequest(w, r, store)
		} else {
			log.Printf("Got unexpected Host value %q\n", r.Host)
			w.WriteHeader(404)
		}
	}
}
