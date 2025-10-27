package main

import (
	"bytes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"log"
	"os"
)

type Game struct{}

type guy struct {
	x, y   float64
	sprite *ebiten.Image
}

var tom = guy{
	x: 5,
	y: 5,
}

func (g *Game) Update() error {
	move := 10.0
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		tom.y -= move
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		tom.y += move
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		tom.x -= move
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		tom.x += move
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tom.x, tom.y)
	op.GeoM.Scale(100.0/float64(tom.sprite.Bounds().Dx()), 100.0/float64(tom.sprite.Bounds().Dy()))
	screen.DrawImage(tom.sprite, op)
	ebitenutil.DebugPrint(screen, "Hello, World!")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	beet, _ := os.ReadFile("cmd/web/img/img.png")
	bert, _, _ := image.Decode(bytes.NewReader(beet))
	tom.sprite = ebiten.NewImageFromImage(bert)

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
