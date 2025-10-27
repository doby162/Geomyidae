package main

import (
	"bytes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"image"
	"log"
	"os"
)

type Game struct{}

type guy struct {
	x, y       float64
	sprite     *ebiten.Image
	jumpFrames int
	canJump    bool
	// socket?
}

var tom = guy{
	x: 5,
	y: 5,
}
var otherPlayers = []guy{}
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
	if tom.y < 200 {
		tom.y += 10
	} else {
		// on the ground
		tom.canJump = true
	}

	newPos := struct {
		x, y float64
	}{tom.x, tom.y}

	if newPos != prevPos {
		// tom.socket.Update() or whatever
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tom.x, tom.y)
	screen.DrawImage(tom.sprite, op)
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	beet, _ := os.ReadFile("assets/img/placeholderSprite.png")
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	tom.sprite = ebiten.NewImageFromImage(bert)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
