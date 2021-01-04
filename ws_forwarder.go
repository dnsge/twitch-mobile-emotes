package tme

import (
	"context"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/session"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	twitchIrcWsURL = "wss://irc-ws.chat.twitch.tv:443"
)

func connectToTwitchIrc(ctx context.Context) (*websocket.Conn, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, twitchIrcWsURL, http.Header{})
	return conn, err
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WsForwarder struct {
	emoteStore  *emotes.EmoteStore
	imageCache  *emotes.ImageFileCache
	includeGifs bool
	ctx         context.Context
}

func NewWsForwarder(store *emotes.EmoteStore, imageCache *emotes.ImageFileCache, includeGifs bool, ctx context.Context) *WsForwarder {
	return &WsForwarder{
		emoteStore:  store,
		imageCache:  imageCache,
		includeGifs: includeGifs,
		ctx:         ctx,
	}
}

func (f *WsForwarder) HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	twitchConn, err := connectToTwitchIrc(f.ctx)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Client connected")
	session.RunWsSession(conn, twitchConn, f.emoteStore, f.imageCache, f.includeGifs)
	log.Println("Client disconnected")
}
