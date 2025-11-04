package main

import (
	"log"
	"strings"

	cp "github.com/jakecoffman/cp/v2"
)

func main() {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{Y: space.Gravity().Y})

	for rowIndex, row := range strings.Split(scene, "\n") {
		for colIndex, col := range row {
			if col == '1' {
				CreateStaticTile(0.5, float64(colIndex)+0.5, float64(rowIndex)+0.5, space)
			}
		}
	}

	CreatePlayerCollider(0.5, 3, 3, 1, space)
	count := 0
	for {
		count++
		if count > 2000 {
			log.Println("adding a player")
			CreatePlayerCollider(0.5, 3, 3, 1, space)
		}
		space.Step(0.1)
		if count > 2000 {
			log.Println("added a player and got through tick")
			count = 0
		}
	}
}

const scene = "11111111111111111111111111111\n" +
	"100000000000000000010000000000\n" +
	"10000000000000000001000000001\n" +
	"10000000000000000001000000001\n" +
	"10000000000000111001000000001\n" +
	"10000000000000000000000000001\n" +
	"10001110000000000000000000001\n" +
	"10000000000000000001000000001\n" +
	"10000000000011100001000000001\n" +
	"10000000000000000001000000001\n" +
	"10000111000000000001000000001\n" +
	"10000000000000000001000000001\n" +
	"11111111111111111111111001111\n" +
	"10000000000000000000000000001\n" +
	"10000000000000000000000000001\n" +
	"10000000000000000000011000001\n" +
	"10000000000000000000000000001\n" +
	"11111111111111111111111111111"

func CreateStaticTile(halfSize, centerX, centerY float64, space *cp.Space) *cp.Body {
	body := cp.NewBody(1, 1)

	shape := cp.NewBox(body, halfSize*2, halfSize*2, 0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: centerX, Y: centerY})

	space.AddShape(shape)
	space.AddBody(body)

	return body
}
func CreatePlayerCollider(halfSize, centerX, centerY float64, fixedRotation uint8, space *cp.Space) *cp.Body {
	body := cp.NewBody(1, 1)

	shape := cp.NewBox(body, halfSize*2, halfSize*2, 0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: centerX, Y: centerY})

	space.AddShape(shape)
	space.AddBody(body)

	return body
}
