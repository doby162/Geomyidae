package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// copy pasta from websocket example code
// not fully clear on what this actually does for us
func handleChannels(done chan struct{}, interrupt chan os.Signal, c WSConn) {
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			slog.Debug("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				slog.Debug("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
