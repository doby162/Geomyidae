package main

// WSConn is a minimal WebSocket-like interface used by the client code.
// This is implemented by a thin wrapper around *websocket.Conn for native builds
// and by a JS/WebSocket wrapper for wasm builds.
type WSConn interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(messageType int, data []byte) error
	Close() error
}
