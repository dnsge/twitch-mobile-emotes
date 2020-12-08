package tme

import (
	"context"
	"fmt"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
	"sync"
	"unicode/utf8"
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
	emoteStore *emotes.EmoteStore
	userIDs    map[string]string
	ctx        context.Context
}

func NewWsForwarder(store *emotes.EmoteStore, ctx context.Context) *WsForwarder {
	return &WsForwarder{
		emoteStore: store,
		userIDs:    make(map[string]string),
		ctx:        ctx,
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

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() { // read from proxy and write to twitch
		defer wg.Done()
		for {
			m, p, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("proxy read error: %v\n", err)
				}
				twitchConn.Close()
				return
			}

			if err = twitchConn.WriteMessage(m, p); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("twitch conn write error: %v\n", err)
				}
				conn.Close()
				return
			}
		}
	}()

	go func() { // read from twitch and write to proxy
		defer wg.Done()
		for {
			mt, p, err := twitchConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("twitch conn read error: %v\n", err)
				}
				conn.Close()
				return
			}

			for _, line := range strings.Split(string(p), "\r\n") {
				if line == "" {
					continue
				}

				sent, e := f.handleMessage(conn, mt, line+"\r\n")
				if !sent {
					if err = conn.WriteMessage(mt, p); err != nil {
						if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
							log.Printf("proxy write error: %v\n", err)
						}
						twitchConn.Close()
						return
					}
				}

				if e != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
						log.Printf("handle message: %v\n", err)
					}
					twitchConn.Close()
					return
				}
			}
		}
	}()

	wg.Wait()
}

// returns whether the message was passed on and an error
func (f *WsForwarder) handleMessage(conn *websocket.Conn, mt int, body string) (bool, error) {
	msg, err := irc.ParseMessage(body)
	if err != nil {
		return false, fmt.Errorf("parse message: %w", err)
	}

	if msg.Command == "PRIVMSG" || msg.Command == "USERNOTICE" {
		channelName := msg.Params[0]
		channelID, found := f.userIDs[channelName]
		if !found {
			if tag, ok := msg.Tags.GetTag("room-id"); ok { // fallback to room-id
				channelID = tag
			} else {
				return false, fmt.Errorf("unmapped user id for channel %q", channelName)
			}
		}

		if err := f.injectThirdPartyEmotes(msg, channelID); err != nil {
			return false, fmt.Errorf("inject emotes: %w", err)
		}

		modified := msg.String() + "\r\n"
		if err = conn.WriteMessage(mt, []byte(modified)); err != nil {
			return false, err
		} else {
			return true, nil // successfully modified and sent message
		}
	} else if msg.Command == "ROOMSTATE" {
		channelID := string(msg.Tags["room-id"])
		f.userIDs[msg.Params[0]] = channelID

		if err := f.emoteStore.LoadIfNotLoaded(channelID); err != nil {
			return false, fmt.Errorf("load channel: %w", err)
		}
	}

	return false, nil
}

func (f *WsForwarder) injectThirdPartyEmotes(msg *irc.Message, channelID string) error {
	messageBody := msg.Trailing()
	emoteTag, err := irc.NewEmoteTag(msg.Tags["emotes"])
	if err != nil {
		return err
	}

	i := 0
	for _, word := range strings.Split(messageBody, " ") {
		wordLen := utf8.RuneCountInString(word) // UTF-8 so emojis don't mess up
		if e, found := f.emoteStore.GetEmoteFromWord(word, channelID); found {
			emoteTag.Add(e.LetterCode()+e.EmoteID(), [2]int{i, i + wordLen - 1})
		}
		i += wordLen + 1
	}

	msg.Tags["emotes"] = emoteTag.TagValue()
	return nil
}
