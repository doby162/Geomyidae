package main

import (
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"encoding/json"
	"time"

	"github.com/jakecoffman/cp/v2"

	assets "Geomyidae"
	"Geomyidae/internal/shared_structs"
	tiled "Geomyidae/internal/tiled"
	"log"
)

var physics *cp.Space
var prevTime time.Time

type WorldData struct {
	Name    string                      `json:"name"`
	Objects []shared_structs.GameObject `json:"objects"`
}

var playerList *player.List
var staticObjectList []*cp.Body
var dynamicObjectList []*cp.Body

func main() {
	// Import tile data
	tileByteInput, err := assets.FS.ReadFile("assets/tiled/test-one.tmx")
	if err != nil {
		log.Fatal(err)
	}
	tileData := tiled.GetTileData(tileByteInput)

	// instantiate state
	physics = cp.NewSpace()
	playerList = player.NewList(physics)

	for _, td := range tileData {
		if td.ID == 0 {
			continue // empty tile
		}
		body := cp.NewStaticBody()
		shape := cp.NewBox(body, 1, 1, 0)
		shape.SetElasticity(0.25)
		shape.SetDensity(0.5)
		shape.SetFriction(1.0)
		body.AddShape(shape)
		body.SetPosition(cp.Vector{X: float64(td.Col) + 0.5, Y: float64(td.Row) + 0.5})

		physics.AddShape(shape)
		staticObjectList = append(staticObjectList, body)
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

		data := collectWorldState(true)

		for sock, _ := range hub.Clients {
			data.Name = sock.Player.Name
			msg, _ := json.Marshal(data)
			sock.Send <- msg
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func collectWorldState(includeStaticAndAsleep bool) *WorldData {
	data := WorldData{}
	if includeStaticAndAsleep {
		for _, object := range staticObjectList {
			pos := object.Position()
			x, y := pos.X, pos.Y
			data.Objects = append(data.Objects, shared_structs.GameObject{
				X:       x,
				Y:       y,
				Sprite:  "platformerPack_industrial",
				OffsetX: 0,
				OffsetY: 0,
				Width:   64,
				Height:  64,
				Name:    "",
				Angle:   object.Angle(),
			})
		}
	}
	for _, object := range dynamicObjectList {
		if object.IsSleeping() && !includeStaticAndAsleep {
			continue
		}
		pos := object.Position()
		x, y := pos.X, pos.Y
		data.Objects = append(data.Objects, shared_structs.GameObject{
			X:       x,
			Y:       y,
			Sprite:  "platformerPack_industrial",
			OffsetX: 0,
			OffsetY: 0,
			Width:   64,
			Height:  64,
			Name:    "",
			Angle:   object.Angle(),
		})
	}
	for _, networkPlayer := range playerList.Players {
		pos := networkPlayer.Body.Position()
		x, y := pos.X, pos.Y
		data.Objects = append(data.Objects, shared_structs.GameObject{
			X:       x,
			Y:       y,
			Sprite:  "spaceShooterRedux",
			OffsetX: 325,
			OffsetY: 0,
			Width:   98,
			Height:  75,
			Name:    networkPlayer.Name,
			Angle:   networkPlayer.Body.Angle(),
		})
	}
	return &WorldData{}
}
