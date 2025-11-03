package main

import (
	"Geomyidae/cmd/server/box"
	"Geomyidae/cmd/server/player"
	"Geomyidae/cmd/server/sock_server"
	"encoding/json"
	"strings"
	"time"
)

// for this refactor I need to take all the physics logic and put it into the server
// the front end will use ebiten and the backend will use box2d
// socket messages will no longer be braodcast, rather they will set values that will be referenced
// by the game loop on the server
// regularly scheduled updates will be sent out to clients
// this is basically a thin client pattern where the client consists of a screen and a gamepad
// but the game fully takes place on the server

var physics box.Physics
var prevTime time.Time

type GameObject struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Sprite string  `json:"sprite"`
}
type WorldData struct {
	Objects []GameObject `json:"objects"`
}

func main() {
	// instantiate state
	physics = box.B2New(9.8)
	playerList := player.NewList(physics)
	var objectList []box.Body

	tile := box.BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}
	for rowIndex, row := range strings.Split(scene01, "\n") {
		for colIndex, col := range row {
			if col == '1' {
				objectList = append(objectList, physics.CreateStaticTile(0.5, float64(colIndex)+0.5, float64(rowIndex)+0.5, tile))
			}
		}
	}
	// kick off socket server
	hub := sock_server.Api(playerList)

	prevTime = time.Now()
	for {
		playerList.WriteAccess.Lock()
		for _, networkPlayer := range playerList.Players {
			networkPlayer.ApplyKeys()
		}
		deltaTime := time.Now().Sub(prevTime).Seconds()
		prevTime = time.Now()
		physics.Step(deltaTime, 1)
		//playerList.WriteAccess.Unlock()
		data := WorldData{}
		for _, object := range objectList {
			x, y := object.Position()
			data.Objects = append(data.Objects, GameObject{x, y, "tile_01"})
		}
		for _, networkPlayer := range playerList.Players {
			x, y := networkPlayer.Body.Position()
			data.Objects = append(data.Objects, GameObject{x, y, networkPlayer.Sprite})
		}
		playerList.WriteAccess.Unlock()
		msg, _ := json.Marshal(data)
		hub.Broadcast <- msg
		time.Sleep(50 * time.Millisecond)
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
