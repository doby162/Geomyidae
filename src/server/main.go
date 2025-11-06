package main

import (
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"

	assets "Geomyidae"
	"Geomyidae/internal/shared_structs"
	tiled "Geomyidae/internal/tiled"
	"log"
)

var physics *cp.Space
var prevTime time.Time

var playerList *player.List
var staticObjectList []*shared_structs.GameObject
var dynamicObjectList []*shared_structs.GameObject

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

		pos := body.Position()
		obj := shared_structs.GameObject{
			X:                    pos.X,
			Y:                    pos.Y,
			Sprite:               "platformerPack_industrial",
			SpriteOffsetX:        0,
			SpriteOffsetY:        0,
			SpriteWidth:          64,
			SpriteHeight:         64,
			SpriteFlipHorizontal: td.SpriteFlipHorizontal,
			SpriteFlipVertical:   td.SpriteFlipVertical,
			SpriteFlipDiagonal:   td.SpriteFlipDiagonal,
			Angle:                body.Angle(),
			UUID:                 uuid.New().String(),
			Body:                 body,
			Shape:                shape,
		}

		physics.AddShape(shape)
		staticObjectList = append(staticObjectList, &obj)
		physics.AddBody(body)
	}

	// kick off socket server
	hub := sock_server.Api(playerList)

	prevTime = time.Now()
	for {
		playerList.WriteAccess.Lock()
		deltaTime := time.Now().Sub(prevTime).Seconds()
		for _, networkPlayer := range playerList.Players {
			networkPlayer.ApplyKeys(deltaTime)
			if networkPlayer.ShootFlag {
				log.Println(networkPlayer.Name)
				networkPlayer.ShootFlag = false
				body := cp.NewBody(1, 1)
				shape := cp.NewCircle(body, 0.125, cp.Vector{X: 0, Y: 0})
				shape.SetElasticity(0.25)
				shape.SetDensity(0.5)
				shape.SetFriction(1.0)
				body.AddShape(shape)
				pos := networkPlayer.Body.Position()
				x := pos.X
				y := pos.Y
				angle := networkPlayer.Body.Angle()
				thrust := 40.0
				offset := 1.0
				x = x + math.Sin(angle)*offset
				y = y + math.Cos(angle)*(offset*-1)
				body.SetVelocity(math.Sin(angle)*thrust, math.Cos(angle)*(-1*thrust))
				body.SetPosition(cp.Vector{X: x, Y: y})

				physics.AddBody(body)
				physics.AddShape(shape)

				dynamicObjectList = append(dynamicObjectList, &shared_structs.GameObject{
					X:             x,
					Y:             y,
					Sprite:        "spaceShooterRedux",
					SpriteOffsetX: 0,
					SpriteOffsetY: 0,
					SpriteWidth:   16,
					SpriteHeight:  16,
					Angle:         body.Angle(),
					UUID:          uuid.New().String(),
					Body:          body,
					Shape:         shape,
				})
			}
		}
		prevTime = time.Now()
		physics.Step(deltaTime)
		playerList.WriteAccess.Unlock()

		data := collectWorldState()

		for sock, _ := range hub.Clients {
			data.Name = sock.Player.Name
			msg, _ := json.Marshal(data)
			sock.Send <- msg
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func collectWorldState() *shared_structs.WorldData {
	data := shared_structs.WorldData{}
	includeStaticAndAsleep := false
	for _, networkPlayer := range playerList.Players {
		if networkPlayer.NeedsStatics {
			log.Println("full packet")
			includeStaticAndAsleep = true
			networkPlayer.NeedsStatics = false
		}
		pos := networkPlayer.Body.Position()
		x, y := pos.X, pos.Y
		data.Objects = append(data.Objects, shared_structs.GameObject{
			X:             x,
			Y:             y,
			Sprite:        "spaceShooterRedux",
			SpriteOffsetX: 325,
			SpriteOffsetY: 0,
			SpriteWidth:   98,
			SpriteHeight:  75,
			Angle:         networkPlayer.Body.Angle(),
			UUID:          networkPlayer.Name,
		})
	}
	if includeStaticAndAsleep {
		for _, object := range staticObjectList {
			data.Objects = append(data.Objects, *object)
		}
	}
	for _, object := range dynamicObjectList {
		if object.Body.IsSleeping() && !includeStaticAndAsleep {
			continue
		}
		pos := object.Body.Position()
		x, y := pos.X, pos.Y
		object.X = x
		object.Y = y
		object.Angle = object.Body.Angle()
		data.Objects = append(data.Objects, *object)
	}
	return &data
}
