package main

import (
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"encoding/json"
	"github.com/jakecoffman/cp/v2"
	"time"

	assets "Geomyidae"
	tiled "Geomyidae/internal/tiled"
	"fmt"
	"log"
)

var physics *cp.Space
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
	physics = cp.NewSpace()
	playerList := player.NewList(physics)
	var objectList []*cp.Body

	for _, td := range tileData {
		if td.ID == 0 {
			continue // empty tile
		}
		fmt.Println("Creating tile at row", td.Row, "col", td.Col)
		body := cp.NewBody(1, 1)
		shape := cp.NewBox(body, 1, 1, 0)
		shape.SetElasticity(0.25)
		shape.SetDensity(0.5)
		shape.SetFriction(1.0)
		body.AddShape(shape)
		body.SetPosition(cp.Vector{X: float64(td.Col) + 0.5, Y: float64(td.Row) + 0.5})

		physics.AddShape(shape)
		objectList = append(objectList, body)
		physics.AddBody(body)
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
		physics.Step(deltaTime)
		playerList.WriteAccess.Unlock()
		data := WorldData{}
		for _, object := range objectList {
			pos := object.Position()
			x, y := pos.X, pos.Y
			data.Objects = append(data.Objects, GameObject{x, y, "tile_01", ""})
		}
		for _, networkPlayer := range playerList.Players {
			pos := networkPlayer.Body.Position()
			x, y := pos.X, pos.Y
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
