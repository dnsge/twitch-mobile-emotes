package session

import (
	"github.com/gorilla/websocket"
	"io"
)

type MessageModifier func(data []byte, writer io.Writer) error

func makeCloseMessageFromError(err error) []byte {
	if e, ok := err.(*websocket.CloseError); ok && e.Code != websocket.CloseNoStatusReceived {
		return websocket.FormatCloseMessage(e.Code, e.Text)
	} else {
		return websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing")
	}
}

func writeWithModifier(conn WsConn, msgType int, msg []byte, modifier MessageModifier) error {
	connWriter, err := conn.NextWriter(msgType)
	if err != nil {
		return err
	}

	if err := modifier(msg, connWriter); err != nil {
		return err
	}

	return connWriter.Close()
}

// proxyConnections copies websocket messages from src to dst, calling modifier on messages.
//
// Errors are returned through errC. src is closed before returning.
func proxyConnections(dst, src WsConn, errC chan<- error, modifier MessageModifier) {
	defer src.Close()
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			closeMsg := makeCloseMessageFromError(err)
			_ = dst.WriteMessage(websocket.CloseMessage, closeMsg)
			errC <- err
			break
		}

		var writeError error
		if modifier != nil {
			writeError = writeWithModifier(dst, msgType, msg, modifier)
		} else {
			writeError = dst.WriteMessage(msgType, msg)
		}

		if writeError != nil {
			closeMsg := makeCloseMessageFromError(err)
			_ = src.WriteMessage(websocket.CloseMessage, closeMsg)
			errC <- err
			break
		}
	}
}
