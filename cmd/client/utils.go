package main

import (
	"log/slog"
	"math/rand"
	"os"
	"time"

	higher_order "github.com/doby162/go-higher-order"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// key checks take a function to run if the key is held
func checkKey(checkKey ebiten.Key, fn func()) {
	if higher_order.AnySlice(heldKeys, func(key ebiten.Key) bool {
		return key == checkKey
	}) {
		fn()
	}
}
func handleKeyState() {
	// keys are added to held when pressed.
	// keys are removed when released
	heldKeys = inpututil.AppendPressedKeys(heldKeys)
	releasedKeys = inpututil.AppendJustReleasedKeys([]ebiten.Key{})
	for _, key := range releasedKeys {
		heldKeys = higher_order.FilterSlice(heldKeys, func(key2 ebiten.Key) bool {
			return key != key2
		})
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// copy pasta from websocket example code
// not fully clear on what this actually does for us
func handleChannels(done chan struct{}, interrupt chan os.Signal, c *websocket.Conn) {
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
