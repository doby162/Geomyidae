package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"sync"

	assets "Geomyidae"

	higher_order "github.com/doby162/go-higher-order"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 1280
	screenHeight = 832
	tileSize     = 64
	tileHalf     = 32
)

var sprites map[string]*ebiten.Image

type Game struct{}

type GameObject struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Sprite string  `json:"sprite"`
	Name   string  `json:"name"`
}
type WorldData struct {
	Objects []GameObject `json:"objects"`
	Name    string       `json:"name"`
}
type keysStruct struct {
	Keys []string `json:"keys"`
}

var heldKeys []ebiten.Key
var world WorldData
var mu sync.Mutex

func (g *Game) Update() error {
	tom := higher_order.FilterSlice(world.Objects, func(o GameObject) bool {
		return world.Name == o.Name
	})
	if len(tom) == 1 {
		cameraX = (tom[0].X * tileSize) - screenWidth/2
		cameraY = (tom[0].Y * tileSize) - screenHeight/2 - (2 * tileSize)
	}

	handleKeyState()

	msg := keysStruct{}
	for _, ekey := range heldKeys {
		msg.Keys = append(msg.Keys, ekey.String())
	}
	msgBytes, _ := json.Marshal(msg)

	err := socket.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		slog.Error(err.Error())
	}

	return nil
}

var cameraX float64
var cameraY float64
var socket WSConn

func (g *Game) Draw(screen *ebiten.Image) {
	mu.Lock()
	defer mu.Unlock()
	for _, object := range world.Objects {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cameraX, -cameraY)
		op.GeoM.Translate(object.X*tileSize-tileHalf, object.Y*tileSize-tileHalf)
		screen.DrawImage(sprites[object.Sprite], op)
	}

	ebitenutil.DebugPrint(screen, "Camera position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", cameraX, cameraY, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// assets are embedded in package "Geomyidae/assets"

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	beet, _ := os.ReadFile("assets/img/placeholderSprite.png")
	beet, err := assets.FS.ReadFile("cmd/assets/img/placeholderSprite.png")
	if err != nil {
		log.Fatal(err)
	}
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	sprites = make(map[string]*ebiten.Image)
	sprites["player_01"] = ebiten.NewImageFromImage(bert)
	sprites["tile_01"] = ebiten.NewImageFromImage(bert)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	slog.Debug("connecting to %s", u.String())

	conn, err := DialWS(u.String())
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c WSConn) {
		err := c.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(conn)

	socket = conn

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("read:", err)
				return
			}
			slog.Debug("recv: %s", message)
			mu.Lock()
			err = json.Unmarshal(message, &world)
			mu.Unlock()
			if err != nil {
				slog.Error("unmarshal:", err)
			}
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go handleChannels(done, interrupt, conn)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, Tom!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
