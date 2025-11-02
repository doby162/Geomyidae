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
}
type WorldData struct {
	Objects []GameObject `json:"objects"`
}
type keysStruct struct {
	Keys []string `json:"keys"`
}

type guy struct {
	x, y       float64
	sprite     *ebiten.Image
	jumpFrames int
	canJump    bool
	name       string
}

var others map[string]*guy

var tom = guy{}
var heldKeys []ebiten.Key
var world WorldData
var mu sync.Mutex

func (g *Game) Update() error {
	//cameraX = tom.x - screenWidth/2
	//cameraY = tom.y - screenHeight/2 - (2 * tileSize)

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
var socket *websocket.Conn

func (g *Game) Draw(screen *ebiten.Image) {
	mu.Lock()
	defer mu.Unlock()
	for _, object := range world.Objects {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cameraX, -cameraY)
		op.GeoM.Translate(object.X*tileSize-tileHalf, object.Y*tileSize-tileHalf)
		screen.DrawImage(sprites[object.Sprite], op)
	}

	ebitenutil.DebugPrint(screen, "Tom's position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", tom.x, tom.y, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	beet, _ := os.ReadFile("assets/img/placeholderSprite.png")
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	sprites = make(map[string]*ebiten.Image)
	sprites["player_01"] = ebiten.NewImageFromImage(bert)
	sprites["tile_01"] = ebiten.NewImageFromImage(bert)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	slog.Debug("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(c)
	socket = c

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
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

	go handleChannels(done, interrupt, c)

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, Tom!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
