package main

import (
	"bufio"
	"github.com/dnsge/twitch-mobile-emotes/session"
	"github.com/gorilla/websocket"
	"io"
	"os"
)

type ConsoleConn struct {
	OnClose func()

	input        io.Reader
	output       io.WriteCloser
	inputScanner *bufio.Scanner
}

func NewConsoleConn() *ConsoleConn {
	input := os.Stdin
	return &ConsoleConn{
		OnClose:      nil,
		input:        input,
		inputScanner: bufio.NewScanner(input),
		output:       &nopCloseWriter{os.Stdout},
	}
}

func (c *ConsoleConn) Close() error {
	if c.OnClose != nil {
		c.OnClose()
	}
	return nil
}

func (c *ConsoleConn) ReadMessage() (int, []byte, error) {
	if !c.inputScanner.Scan() {
		return websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), nil
	}

	return websocket.TextMessage, c.inputScanner.Bytes(), nil
}

func (c *ConsoleConn) WriteMessage(messageType int, data []byte) error {
	_, err := c.output.Write(data)
	return err
}

func (c *ConsoleConn) NextReader() (int, io.Reader, error) {
	return websocket.TextMessage, c.input, nil
}

func (c *ConsoleConn) NextWriter(messageType int) (io.WriteCloser, error) {
	return c.output, nil
}

var _ session.WsConn = &ConsoleConn{}

type nopCloseWriter struct {
	io.Writer
}

func (n *nopCloseWriter) Close() error {
	return nil
}
