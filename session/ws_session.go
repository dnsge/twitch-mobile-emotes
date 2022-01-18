package session

import (
	"bufio"
	"github.com/dnsge/twitch-mobile-emotes/app"
	"github.com/dnsge/twitch-mobile-emotes/emotes"
	"github.com/dnsge/twitch-mobile-emotes/irc"
	"github.com/dnsge/twitch-mobile-emotes/storage"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net"
)

const CRLF = "\r\n"

type WsConn interface {
	Close() error
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	NextReader() (int, io.Reader, error)
	NextWriter(messageType int) (io.WriteCloser, error)
}

func RunWsSession(clientConn, twitchConn WsConn, ctx *app.Context) {
	session := &wsSession{
		config:             ctx.Config,
		clientConn:         clientConn,
		twitchConn:         twitchConn,
		emoteStore:         ctx.EmoteStore,
		imageCache:         ctx.ImageCache,
		settingsRepository: ctx.SettingsRepository,

		defaultIncludeGifs: ctx.Config.IncludeGifs,

		state: &state{
			Username: "",
			UserID:   "",
			Greeted:  false,
		},
		settings: nil,
	}
	session.run()
}

type wsSession struct {
	config             *app.ServerConfig
	clientConn         WsConn
	twitchConn         WsConn
	emoteStore         *emotes.EmoteStore
	imageCache         *emotes.ImageFileCache
	settingsRepository storage.SettingsRepository

	defaultIncludeGifs bool

	state    *state
	settings *storage.Settings
}

type state struct {
	Username string
	UserID   string
	Greeted  bool
}

func (s *wsSession) saveSettings() {
	if s.settingsRepository == nil || s.settings == nil || s.state.UserID == "" {
		return
	}

	go func() {
		if err := s.settingsRepository.Save(s.state.UserID, s.settings); err != nil {
			log.Printf("Error saving settings: %v\n", err)
		}
	}()
}

func (s *wsSession) showGifs() bool {
	if s.settings == nil {
		return s.defaultIncludeGifs
	} else {
		return s.settings.EnableGifEmotes
	}
}

func (s *wsSession) run() {
	twitchChan := make(chan error, 1)
	clientChan := make(chan error, 1)
	go proxyConnections(s.clientConn, s.twitchConn, twitchChan, s.modifyTwitchMessage) // incoming messages from twitch
	go proxyConnections(s.twitchConn, s.clientConn, clientChan, s.modifyClientMessage) // outgoing messages from user

	var err error
	var errorMessage string
	select {
	case err = <-twitchChan:
		errorMessage = "Error proxying from twitch to client: %v\n"
	case err = <-clientChan:
		errorMessage = "Error proxying from client to twitch: %v\n"
	case <-s.config.Context.Done():
		s.clientConn.Close()
		s.twitchConn.Close()
		return
	}

	if closeErr, ok := err.(*websocket.CloseError); !ok || closeErr.Code == websocket.CloseAbnormalClosure {
		log.Printf(errorMessage, err)
	}
}

func (s *wsSession) modifyTwitchMessage(reader io.Reader, writer io.Writer) *RWError {
	// sometimes twitch sends multiple messages at once, so loop over each line
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		msg, err := irc.ParseMessage(line)
		if err != nil {
			log.Printf("parse twitch irc message: %v\n", err)

			// Attempt to write the single message line and continue
			if _, err := writer.Write([]byte(line + CRLF)); err != nil {
				return WriteError(err)
			}
			continue
		}

		modified, err := s.handleTwitchMessage(msg)
		if err != nil {
			log.Printf("handle twitch irc message: %v\n", err)

			// Attempt to write the single message line and continue
			if _, err := writer.Write([]byte(line + CRLF)); err != nil {
				return WriteError(err)
			}
			continue
		}

		var rawMessage []byte
		if modified {
			rawMessage = []byte(msg.String() + CRLF)
		} else {
			rawMessage = []byte(line + CRLF)
		}

		if _, err := writer.Write(rawMessage); err != nil {
			return WriteError(err)
		}
	}

	return ReadError(scanner.Err())
}

func (s *wsSession) modifyClientMessage(reader io.Reader, writer io.Writer) *RWError {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		msg, err := irc.ParseMessage(line)
		if err != nil {
			log.Printf("parse client irc message: %v\n", err)

			// Attempt to write the single message line and continue
			if _, err := writer.Write([]byte(line + CRLF)); err != nil {
				return WriteError(err)
			}
			continue
		}

		passOn, modified, err := s.handleClientMessage(msg)
		if err != nil {
			log.Printf("handle client irc message: %v\n", err)

			// Attempt to write the single message line and continue
			if _, err := writer.Write([]byte(line + CRLF)); err != nil {
				return WriteError(err)
			}
			continue
		}

		if !passOn {
			// Skip writing message
			continue
		}

		var rawMessage []byte
		if modified {
			rawMessage = []byte(msg.String() + CRLF)
		} else {
			rawMessage = []byte(line + CRLF)
		}

		if _, err := writer.Write(rawMessage); err != nil {
			return WriteError(err)
		}
	}

	return ReadError(scanner.Err())
}

// Writes a message to the client connection. Returns whether it succeeded.
func (s *wsSession) writeClientMessage(mt int, msg *irc.Message) bool {
	msgBytes := []byte(msg.String() + CRLF) // add CRLF to end of string
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
	msgBytes := []byte(msg.String() + CRLF) // add CRLF to end of string
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
