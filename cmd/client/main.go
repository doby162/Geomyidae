package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	higher_order "github.com/doby162/go-higher-order"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	b2 "github.com/oliverbestmann/box2d-go"
	"image"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
)

const (
	screenWidth  = 1280
	screenHeight = 832
)

type Game struct{}

type guy struct {
	x, y       float64
	sprite     *ebiten.Image
	jumpFrames int
	canJump    bool
	sock       *websocket.Conn
	name       string
	body       Body
}

var others []*guy

var tom = guy{}
var heldKeys []ebiten.Key
var releasedKeys []ebiten.Key
var move = 10.0
var jump = 25.0

func (g *Game) Update() error {
	physics.Step(1, 1) // needs delta time
	prevPos := struct {
		x, y float64
	}{tom.x, tom.y}

	x, y := tom.body.Position()
	tom.x = x * 64
	tom.y = y * 64

	handleKeyState()

	checkKey(ebiten.KeyA, func() {
		tom.body.ApplyForce(b2.Vec2{
			X: -1,
			Y: 0,
		})
	})
	checkKey(ebiten.KeyD, func() {
		tom.body.ApplyForce(b2.Vec2{
			X: 1,
			Y: 0,
		})
	})
	checkKey(ebiten.KeyW, func() {
		tom.body.ApplyForce(b2.Vec2{
			X: 0,
			Y: -5,
		})
	})

	newPos := struct {
		x, y float64
	}{tom.x, tom.y}

	if newPos != prevPos {
		msg := fmt.Sprintf("{\"x\": %v, \"y\": %v, \"name\": \"%v\"}", newPos.x, newPos.y, tom.name)
		err := tom.sock.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

type updateMsg struct {
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(64, 64)
	physics.Draw(screen, op.GeoM)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tom.x-32, tom.y-32)
	screen.DrawImage(tom.sprite, op)
	for _, ourGuy := range others {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ourGuy.x, ourGuy.y)
		screen.DrawImage(ourGuy.sprite, op)
	}
	ebitenutil.DebugPrint(screen, "Tom's position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", tom.x, tom.y, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

var physics Physics

func main() {
	//slog.SetLogLoggerLevel(slog.LevelDebug) // uncomment for verbose logs
	beet, _ := os.ReadFile("assets/img/placeholderSprite.png")
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	tom.sprite = ebiten.NewImageFromImage(bert)
	tom.name = generateRandomString(16)

	physics = b2New(0.5)

	var bodies []Body

	tile := BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}
	box := BodyDef{Elasticity: 0.25, Friction: 0.5, Density: 1}

	for rowIndex, row := range strings.Split(scene01, "\n") {
		for colIndex, col := range row {
			if col == '1' {
				log.Println(float64(32 + (64 * rowIndex)))
				bodies = append(bodies, physics.CreateStaticTile(0.5, float64(colIndex)+0.5, float64(rowIndex)+0.5, tile))
			}
		}
	}

	tom.body = physics.CreateSquare(0.5, 3, 3, box)
	bodies = append(bodies, tom.body)

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

	tom.sock = c

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				slog.Error("read:", err)
				return
			}
			m := updateMsg{}
			var ourGuy *guy
			err = json.Unmarshal(message, &m)
			if err != nil {
				slog.Error("unmarshal:", err)
			} else if m.Name == tom.name {
			} else if higher_order.AnySlice(others, func(g *guy) bool {
				if g.name == m.Name {
					ourGuy = g
					return true
				}
				return false
			}) { // if we have the guy already
				slog.Debug("found our guy")
				ourGuy.x = m.X
				ourGuy.y = m.Y
				ourGuy.body.SetPosition((m.X/64)+32, (m.Y/64)+32)
			} else { //  if we  have to make a new guy
				slog.Debug("make a new guy")
				bod := physics.CreateSquare(0.5, (m.X/64)+32, (m.Y/64)+32, box)
				bodies = append(bodies, bod) // we don't actually do anything with this yet
				ourGuy = &guy{x: m.X, y: m.Y, sprite: tom.sprite, name: m.Name, body: bod}
				others = append(others, ourGuy)
			}
			slog.Debug("recv: %s", message)
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
