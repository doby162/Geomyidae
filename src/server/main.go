package main

import (
	"Geomyidae/server/box"
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"encoding/json"
	"time"

	assets "Geomyidae"
	tiled "Geomyidae/internal/tiled"
	"fmt"
	"log"
)

var physics box.Physics
var prevTime time.Time

type GameObject struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Sprite string  `json:"sprite"`
	Name   string  `json:"name"`
}
type WorldData struct {
	Name    string       `json:"name"`
	Objects []GameObject `json:"objects"`
}

func main() {
	// Import tile data
	tileByteInput, err := assets.FS.ReadFile("assets/tiled/test-one.tmx")
	if err != nil {
		log.Fatal(err)
	}
	tileData := tiled.GetTileData(tileByteInput)
	fmt.Println("Total tiles loaded:", len(tileData))

	// instantiate state
	physics = box.B2New(9.8)
	playerList := player.NewList(physics)
	var objectList []box.Body

	tile := box.BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}

	for _, td := range tileData {
		if td.ID == 0 {
			continue // empty tile
		}
		fmt.Println("Creating tile at row", td.Row, "col", td.Col)
		objectList = append(objectList, physics.CreateStaticTile(0.5, float64(td.Col)+0.5, float64(td.Row)+0.5, tile))
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
		playerList.WriteAccess.Unlock()
		data := WorldData{}
		for _, object := range objectList {
			x, y := object.Position()
			data.Objects = append(data.Objects, GameObject{x, y, "tile_01", ""})
		}
		for _, networkPlayer := range playerList.Players {
			x, y := networkPlayer.Body.Position()
			data.Objects = append(data.Objects, GameObject{x, y, networkPlayer.Sprite, networkPlayer.Name})
		}
		for sock, _ := range hub.Clients {
			data.Name = sock.Player.Name
			msg, _ := json.Marshal(data)
			sock.Send <- msg
		}
		time.Sleep(50 * time.Millisecond)
	}
}
