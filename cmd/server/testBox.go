package main

import (
	"Geomyidae/cmd/server/box"
	"strings"
)

func main() {
	tile := box.BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}
	physics := box.B2New(9.8)

	for rowIndex, row := range strings.Split(scene01, "\n") {
		for colIndex, col := range row {
			if col == '1' {
				physics.CreateStaticTile(0.5, float64(colIndex)+0.5, float64(rowIndex)+0.5, tile)
			}
		}
	}
	bd := box.BodyDef{Elasticity: 0.25, Friction: 0.0, Density: 1}

	physics.CreatePlayerCollider(0.5, 3, 3, bd, 1, 0.1)
	count := 0
	for {
		count++
		if count > 200 {
			count = 0
			physics.CreatePlayerCollider(0.5, 3, 3, bd, 1, 0.1)
		}
		physics.Step(0.1, 1)
	}
}

const scene01 = "11111111111111111111111111111\n" +
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
