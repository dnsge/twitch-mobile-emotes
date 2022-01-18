package tme

import (
	"context"
	"github.com/dnsge/twitch-mobile-emotes/app"
	"github.com/dnsge/twitch-mobile-emotes/session"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

const (
	twitchIrcWsURL = "wss://irc-ws.chat.twitch.tv:443"
)

func ConnectToTwitchIrc(ctx context.Context) (*websocket.Conn, error) {
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, twitchIrcWsURL, http.Header{})
	return conn, err
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// trust all origins
		return true
	},
}

type WsForwarder struct {
	ctx *app.Context
}

func NewWsForwarder(ctx *app.Context) *WsForwarder {
	return &WsForwarder{
		ctx: ctx,
	}
}

func (f *WsForwarder) HandleWsConnection(w http.ResponseWriter, r *http.Request) {
	twitchConn, err := ConnectToTwitchIrc(f.ctx.Config.Context)
	if err != nil {
		log.Printf("Failed to connect to Twitch IRC server: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	session.RunWsSession(conn, twitchConn, f.ctx)
}
