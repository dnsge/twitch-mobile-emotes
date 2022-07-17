package session

import (
	"bytes"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
)

type MessageModifier func(reader io.Reader, writer io.Writer) *RWError

func makeCloseMessageFromError(err error) []byte {
	if e, ok := err.(*websocket.CloseError); ok && e.Code != websocket.CloseNoStatusReceived {
		return websocket.FormatCloseMessage(e.Code, e.Text)
	} else {
		return websocket.FormatCloseMessage(websocket.CloseNormalClosure, "closing")
	}
}

func processMessageWithModifier(dst, src WsConn, modifier MessageModifier) *RWError {
	msgType, msgBytes, err := src.ReadMessage()
	if err != nil {
		return ReadError(err)
	}

	connWriter, err := dst.NextWriter(msgType)
	if err != nil {
		return WriteError(err)
	}

	buf := bytes.NewBuffer(msgBytes)
	if err := modifier(buf, connWriter); !err.Ok() {
		return err
	}

	return WriteError(connWriter.Close())
}

// proxyConnections copies websocket messages from src to dst, calling modifier on messages.
//
// Errors are returned through errC. src is closed before returning.
func proxyConnections(dst, src WsConn, errC chan<- error, modifier MessageModifier) {
	defer src.Close()
	for {
		var rwErr *RWError
		if modifier != nil {
			rwErr = processMessageWithModifier(dst, src, modifier)
		} else {
			msgType, msg, err := src.ReadMessage()
			if err != nil {
				rwErr = ReadError(err)
			} else {
				rwErr = WriteError(dst.WriteMessage(msgType, msg))
			}
		}

		if rwErr.Ok() {
			continue
		} else if rwErr.ReadError != nil {
			closeMsg := makeCloseMessageFromError(rwErr.ReadError)
			_ = dst.WriteMessage(websocket.CloseMessage, closeMsg)
			errC <- rwErr.ReadError
			break
		} else if rwErr.WriteError != nil {
			closeMsg := makeCloseMessageFromError(rwErr.WriteError)
			_ = src.WriteMessage(websocket.CloseMessage, closeMsg)
			errC <- rwErr.WriteError
			break
		}
	}
}

type RWError struct {
	ReadError  error
	WriteError error
}

func ReadError(err error) *RWError {
	if err == nil {
		return nil
	}

	return &RWError{
		ReadError:  err,
		WriteError: nil,
	}
}

func WriteError(err error) *RWError {
	if err == nil {
		return nil
	}

	return &RWError{
		ReadError:  nil,
		WriteError: err,
	}
}

func (rw *RWError) Error() string {
	if rw.ReadError != nil && rw.WriteError != nil {
		return fmt.Sprintf("read: %v | write: %v\n", rw.ReadError, rw.WriteError)
	} else if rw.ReadError != nil {
		return fmt.Sprintf("read: %v", rw.ReadError)
	} else if rw.WriteError != nil {
		return fmt.Sprintf("write: %v", rw.WriteError)
	} else {
		return fmt.Sprintf("no error")
	}
}

func (rw *RWError) Ok() bool {
	if rw == nil {
		return true
	}

	return rw.ReadError == nil && rw.WriteError == nil
}
