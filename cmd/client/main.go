package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	higher_order "github.com/doby162/go-higher-order"
	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct{}

type guy struct {
	x, y       float64
	sprite     *ebiten.Image
	jumpFrames int
	canJump    bool
	sock       *websocket.Conn
	name       string
}

var others []*guy

var tom = guy{
	x: 5,
	y: 5,
}
var heldKeys []ebiten.Key
var releasedKeys []ebiten.Key
var move = 10.0
var jump = 25.0

func (g *Game) Update() error {
	prevPos := struct {
		x, y float64
	}{tom.x, tom.y}

	handleKeyState()

	checkKey(ebiten.KeyA, func() { tom.x -= move })
	checkKey(ebiten.KeyD, func() { tom.x += move })
	checkKey(ebiten.KeyW, func() {
		if tom.canJump {
			tom.jumpFrames = 10
			tom.canJump = false
		}
	})

	if tom.jumpFrames > 0 {
		tom.jumpFrames -= 1
		tom.y -= jump
	}

	// please appreciate the gravity of the situation
	if tom.y < screenHeight-70 {
		tom.y += 10
	} else {
		// on the ground
		tom.canJump = true
	}

	newPos := struct {
		x, y float64
	}{tom.x, tom.y}

	if newPos != prevPos {
		msg := fmt.Sprintf("{\"x\": %v, \"y\": %v, \"name\": \"%v\"}", newPos.x, newPos.y, tom.name)
		err := tom.sock.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Print(err.Error())
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
	op.GeoM.Translate(tom.x, tom.y)
	screen.DrawImage(tom.sprite, op)
	for _, ourGuy := range others {
		log.Printf("%+v", ourGuy)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ourGuy.x, ourGuy.y)
		screen.DrawImage(ourGuy.sprite, op)
	}
	ebitenutil.DebugPrint(screen, "Tom's position: "+fmt.Sprintf("%.2f, %.2f", tom.x, tom.y))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	beet, _ := os.ReadFile("assets/img/placeholderSprite.png")
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	tom.sprite = ebiten.NewImageFromImage(bert)
	tom.name = generateRandomString(16)

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	tom.sock = c

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			m := updateMsg{}
			var ourGuy *guy
			err = json.Unmarshal(message, &m)
			if err != nil {
				log.Println("unmarshal:", err)
			} else if m.Name == tom.name {
			} else if higher_order.AnySlice(others, func(g *guy) bool {
				if g.name == m.Name {
					ourGuy = g
					return true
				}
				return false
			}) { // if we have the guy already
				log.Println("found our guy")
				ourGuy.x = m.X
				ourGuy.y = m.Y
			} else { //  if we  have to make a new guy
				log.Println("make a new guy")
				ourGuy = &guy{x: m.X, y: m.Y, sprite: tom.sprite, name: m.Name}
				others = append(others, ourGuy)
			}
			//log.Printf("recv: %s", message)
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-interrupt:
				log.Println("interrupt")

				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	}()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
