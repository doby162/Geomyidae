//go:build !js || !wasm
// +build !js !wasm

package main

import (
	"github.com/gorilla/websocket"
)

// nativeWebSocket wraps *websocket.Conn to implement WSConn for native builds.
type nativeWebSocket struct {
	c *websocket.Conn
}

func (n *nativeWebSocket) ReadMessage() (int, []byte, error) {
	return n.c.ReadMessage()
}

func (n *nativeWebSocket) WriteMessage(messageType int, data []byte) error {
	return n.c.WriteMessage(messageType, data)
}

func (n *nativeWebSocket) Close() error {
	return n.c.Close()
}

// DialWS dials a websocket URL for native builds and returns a WSConn.
func DialWS(u string) (WSConn, error) {
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, err
	}
	return &nativeWebSocket{c: c}, nil
}
