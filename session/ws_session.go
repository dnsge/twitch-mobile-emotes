package session

import (
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"strings"
	"sync"
)

func RunWsSession(clientConn, twitchConn *websocket.Conn, emoteStore *emotes.EmoteStore, imageCache *emotes.ImageFileCache, includeGifs bool) {
	session := &wsSession{
		clientConn:  clientConn,
		twitchConn:  twitchConn,
		emoteStore:  emoteStore,
		imageCache:  imageCache,
		includeGifs: includeGifs,

		greeted: false,
	}
	session.run()
}

type wsSession struct {
	clientConn  *websocket.Conn
	twitchConn  *websocket.Conn
	emoteStore  *emotes.EmoteStore
	imageCache  *emotes.ImageFileCache
	includeGifs bool

	username string
	greeted  bool
}

func (s *wsSession) run() {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go s.clientReadLoop(wg)
	go s.clientWriteLoop(wg)

	wg.Wait()
}

func (s *wsSession) clientReadLoop(wg *sync.WaitGroup) { // read from proxy and write to twitch
	defer wg.Done()
	for {
		mt, p, err := s.clientConn.ReadMessage()
		if err != nil {
			if !isCloseError(err) {
				log.Printf("client conn read error: %v\n", err)
			}
			s.twitchConn.Close()
			return
		}

		for _, line := range strings.Split(string(p), "\r\n") {
			if line == "" {
				continue
			}

			msg, err := irc.ParseMessage(line)
			if err != nil {
				log.Printf("parse irc message: %v\n", err)
				if !s.writeRawTwitchMessage(mt, p) {
					return
				}
				continue
			}

			passOn, modified, err := s.handleClientMessage(msg)
			if err != nil {
				log.Printf("handle irc message: %v\n", err)
				if !s.writeRawTwitchMessage(mt, p) {
					return
				}
				continue
			}

			if !passOn {
				continue
			}

			var success bool
			if modified {
				success = s.writeTwitchMessage(mt, msg)
			} else {
				success = s.writeRawTwitchMessage(mt, p)
			}

			if !success {
				return
			}
		}
	}
}

func (s *wsSession) clientWriteLoop(wg *sync.WaitGroup) { // read from twitch and write to proxy
	defer wg.Done()
	for {
		mt, p, err := s.twitchConn.ReadMessage()
		if err != nil {
			if !isCloseError(err) {
				log.Printf("twitch conn read error: %v\n", err)
			}
			s.clientConn.Close()
			return
		}

		// sometimes twitch sends multiple messages at once, so loop over each line
		for _, line := range strings.Split(string(p), "\r\n") {
			if line == "" {
				continue
			}

			msg, err := irc.ParseMessage(line)
			if err != nil {
				log.Printf("parse irc message: %v\n", err)
				if !s.writeRawClientMessage(mt, p) {
					return
				}
				continue
			}

			modified, err := s.handleTwitchMessage(msg)

			if err != nil {
				log.Printf("handle irc message: %v\n", err)
				if !s.writeRawClientMessage(mt, p) {
					return
				}
				continue
			}

			var success bool
			if modified {
				success = s.writeClientMessage(mt, msg)
			} else {
				success = s.writeRawClientMessage(mt, p)
			}

			if !success {
				return
			}
		}
	}
}

// Writes a message to the client connection. Returns whether it succeeded.
func (s *wsSession) writeClientMessage(mt int, msg *irc.Message) bool {
	msgBytes := []byte(msg.String() + "\r\n") // add CRLF to end of string
	if err := s.clientConn.WriteMessage(mt, msgBytes); err != nil {
		if !isCloseError(err) {
			log.Printf("client conn write error: %v\n", err)
		}
		s.twitchConn.Close()
		return false
	}
	return true
}

// Writes a raw message to the client connection. Returns whether it succeeded.
func (s *wsSession) writeRawClientMessage(mt int, msg []byte) bool {
	if err := s.clientConn.WriteMessage(mt, msg); err != nil { // write original
		if !isCloseError(err) {
			log.Printf("client conn write error: %v\n", err)
		}
		s.twitchConn.Close()
		return false
	}
	return true
}

// Writes a message to the twitch connection. Returns whether it succeeded.
func (s *wsSession) writeTwitchMessage(mt int, msg *irc.Message) bool {
	msgBytes := []byte(msg.String() + "\r\n") // add CRLF to end of string
	if err := s.twitchConn.WriteMessage(mt, msgBytes); err != nil {
		if !isCloseError(err) {
			log.Printf("twitch conn write error: %v\n", err)
		}
		s.clientConn.Close()
		return false
	}
	return true
}

// Writes a raw message to the twitch connection. Returns whether it succeeded.
func (s *wsSession) writeRawTwitchMessage(mt int, msg []byte) bool {
	if err := s.twitchConn.WriteMessage(mt, msg); err != nil { // write original
		if !isCloseError(err) {
			log.Printf("twitch conn write error: %v\n", err)
		}
		s.clientConn.Close()
		return false
	}
	return true
}

func isCloseError(err error) bool {
	_, ok := err.(*websocket.CloseError)
	if ok {
		return true
	}
	_, ok = err.(*net.OpError)
	return ok
}
