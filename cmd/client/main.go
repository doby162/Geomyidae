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
	"strings"
	"time"

	assets "Geomyidae"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	b2 "github.com/oliverbestmann/box2d-go"
)

const (
	screenWidth  = 1280
	screenHeight = 832
	tileSize     = 64
	tileHalf     = 32
)

type Game struct{}

type guy struct {
	x, y       float64
	sprite     *ebiten.Image
	jumpFrames int
	canJump    bool
	sock       WSConn
	name       string
	body       Body
}

var others map[string]*guy

var tom = guy{}
var heldKeys []ebiten.Key
var releasedKeys []ebiten.Key
var move = float32(10.0)
var jump = float32(8.0)
var deltaJump float64

var prevTime time.Time

func (g *Game) Update() error {
	deltaTime := time.Now().Sub(prevTime).Seconds()
	deltaJump += deltaTime
	prevTime = time.Now()
	physics.Step(deltaTime, 1)
	prevPos := struct {
		x, y float64
	}{tom.x, tom.y}

	x, y := tom.body.Position()
	tom.x = x * tileSize
	tom.y = y * tileSize

	cameraX = tom.x - screenWidth/2
	cameraY = tom.y - screenHeight/2 - (2 * tileSize)

	for _, other := range others {
		// very important to do this in the same thread as physics.step to avoid concurrent modification
		other.body.SetPosition(other.x/64, other.y/64)
	}

	handleKeyState()

	checkKey(ebiten.KeyA, func() {
		tom.body.ApplyForce(b2.Vec2{
			X: -move,
			Y: 0,
		})
	})
	checkKey(ebiten.KeyD, func() {
		tom.body.ApplyForce(b2.Vec2{
			X: move,
			Y: 0,
		})
	})
	checkKey(ebiten.KeyW, func() {
		if deltaJump > 2 {
			deltaJump = 0
			// this should really be an impulse not a force.
			tom.body.ApplyImpulse(b2.Vec2{
				X: 0,
				Y: -jump,
			})
		}
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

var cameraX float64
var cameraY float64

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(tileSize, tileSize)
	op.GeoM.Translate(-cameraX, -cameraY)
	physics.Draw(screen, op.GeoM)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cameraX, -cameraY)
	op.GeoM.Translate(tom.x-tileHalf, tom.y-tileHalf)
	screen.DrawImage(tom.sprite, op)
	for _, ourGuy := range others {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-cameraX, -cameraY)
		op.GeoM.Translate(ourGuy.x-tileHalf, ourGuy.y-tileHalf)
		screen.DrawImage(ourGuy.sprite, op)
	}
	ebitenutil.DebugPrint(screen, "Tom's position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", tom.x, tom.y, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

var physics Physics

// assets are embedded in package "Geomyidae/assets"

func main() {
	//slog.SetLogLoggerLevel(slog.LevelDebug) // uncomment for verbose logs
	beet, err := assets.FS.ReadFile("cmd/assets/img/placeholderSprite.png")
	if err != nil {
		log.Fatal(err)
	}
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	tom.sprite = ebiten.NewImageFromImage(bert)
	tom.name = generateRandomString(16)

	others = make(map[string]*guy) // initialize map
	prevTime = time.Now()

	physics = b2New(9.8)

	var bodies []Body

	tile := BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}
	box := BodyDef{Elasticity: 0.25, Friction: 0.0, Density: 1}

	for rowIndex, row := range strings.Split(scene01, "\n") {
		for colIndex, col := range row {
			if col == '1' {
				bodies = append(bodies, physics.CreateStaticTile(0.5, float64(colIndex)+0.5, float64(rowIndex)+0.5, tile))
			}
		}
	}

	tom.body = physics.CreatePlayerCollider(0.5, 3, 3, box, 1, 0.1)
	bodies = append(bodies, tom.body)

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

	tom.sock = conn

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				slog.Error("read:", err)
				return
			}
			m := updateMsg{}
			err = json.Unmarshal(message, &m)
			if err != nil {
				slog.Error("unmarshal:", err)
			} else if m.Name == tom.name {
			} else if others[m.Name] != nil { // if we have the guy already
				slog.Debug("found our guy")
				others[m.Name].x = m.X
				others[m.Name].y = m.Y
			} else { //  if we  have to make a new guy
				slog.Debug("make a new guy")
				bod := physics.CreateNetworkCollider(0.5, 3, 3, box, 1, 0.1)
				bodies = append(bodies, bod) // we don't actually do anything with this yet
				others[m.Name] = &guy{x: m.X, y: m.Y, sprite: tom.sprite, name: m.Name, body: bod}
			}
			slog.Debug("recv: %s", message)
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
