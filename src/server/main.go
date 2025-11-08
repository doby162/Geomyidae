package main

import (
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"

	assets "Geomyidae"
	"Geomyidae/internal/shared_structs"
	tiled "Geomyidae/internal/tiled"
	"log"
)

var physics *cp.Space

const metersToPixels = 64

var prevTime time.Time

// players is a subset of simulationObjects. Every value it contains is a duplicate
var players *player.List
var simulationObjects []shared_structs.HasBehavior

// abuse closures to allow remotely updating state without complicating this pointer situation
func apOb(networkPlayer *player.NetworkPlayer) {
	simulationObjects = append(simulationObjects, networkPlayer)
}

func main() {
	// Import tile data
	tileByteInput, err := assets.FS.ReadFile("assets/tiled/test-one.tmx")
	if err != nil {
		log.Fatal(err)
	}
	tileData := tiled.GetTileData(tileByteInput)

	// instantiate chipmunk
	physics = cp.NewSpace()
	players = player.NewList(physics, apOb)

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

		obj := tile{&shared_structs.GameObject{
			Sprite:               td.Sprite,
			SpriteOffsetX:        td.SpriteOffsetX,
			SpriteOffsetY:        td.SpriteOffsetY,
			SpriteWidth:          td.SpriteWidth,
			SpriteHeight:         td.SpriteHeight,
			SpriteFlipHorizontal: td.SpriteFlipHorizontal,
			SpriteFlipVertical:   td.SpriteFlipVertical,
			SpriteFlipDiagonal:   td.SpriteFlipDiagonal,
			Angle:                body.Angle(),
			UUID:                 uuid.New().String(),
			Body:                 body,
			Shape:                shape,
			IsStatic:             true,
		}}

		physics.AddShape(shape)
		simulationObjects = append(simulationObjects, obj)
		physics.AddBody(body)
	}

	// kick off socket server
	hub := sock_server.Api(players)

	prevTime = time.Now()
	for {
		includeStaticAndAsleep := false
		players.WriteAccess.Lock()
		deltaTime := time.Now().Sub(prevTime).Seconds()

		for _, obj := range simulationObjects {
			obj.ApplyBehavior(deltaTime)
			gameObj := obj.GetObject()
			if gameObj.NeedsStatics {
				log.Println("full packet")
				includeStaticAndAsleep = true
				gameObj.NeedsStatics = false
			}
			if gameObj.ShootFlag {
				gameObj.ShootFlag = false
				body := cp.NewBody(1, 1)
				shape := cp.NewCircle(body, 0.125, cp.Vector{X: 0, Y: 0})
				shape.SetElasticity(0.25)
				shape.SetDensity(0.5)
				shape.SetFriction(1.0)
				body.AddShape(shape)
				pos := gameObj.Body.Position()
				x := pos.X
				y := pos.Y
				angle := gameObj.Body.Angle()
				thrust := 40.0
				offset := 1.0
				x = x + math.Sin(angle)*offset
				y = y + math.Cos(angle)*(offset*-1)
				body.SetVelocity(math.Sin(angle)*thrust, math.Cos(angle)*(-1*thrust))
				body.SetPosition(cp.Vector{X: x, Y: y})

				physics.AddBody(body)
				physics.AddShape(shape)

				simulationObjects = append(simulationObjects, bullet{
					expirationDate: time.Now().Add(time.Second * 5),
					GameObject: &shared_structs.GameObject{
						Sprite:        "spaceShooterRedux",
						SpriteOffsetX: 0,
						SpriteOffsetY: 0,
						SpriteWidth:   16,
						SpriteHeight:  16,
						Angle:         body.Angle(),
						UUID:          uuid.New().String(),
						Body:          body,
						Shape:         shape,
					},
				})
			}
		}

		prevTime = time.Now()
		physics.Step(deltaTime)
		players.WriteAccess.Unlock()

		for _, obj := range simulationObjects {
			gameObj := obj.GetObject()
			pos := gameObj.Body.Position()
			gameObj.X = pos.X * metersToPixels
			gameObj.Y = pos.Y * metersToPixels
			gameObj.Angle = gameObj.Body.Angle()
		}

		data := collectWorldState(includeStaticAndAsleep)

		for sock, _ := range hub.Clients {
			data.Name = sock.Player.UUID
			testUUID = sock.Player.UUID
			msg, _ := json.Marshal(data)
			sock.Send <- msg
		}
		time.Sleep(20 * time.Millisecond)
	}
}

var testUUID = ""

func collectWorldState(includeStaticAndAsleep bool) *shared_structs.WorldData {
	data := shared_structs.WorldData{}
	var removed []*shared_structs.GameObject
	removedIndexes := []int{}

	for index, obj := range simulationObjects {
		gameObj := obj.GetObject()
		if gameObj.Delete {
			physics.RemoveBody(gameObj.Body)
			physics.RemoveShape(gameObj.Shape)
			removed = append(removed, gameObj)
			removedIndexes = append(removedIndexes, index)
		}
		if !includeStaticAndAsleep && gameObj.IsStatic {
			continue
		}
		if !includeStaticAndAsleep && gameObj.Body.IsSleeping() {
			continue
		}
		data.Objects = append(data.Objects, *obj.GetObject())
	}
	// it doesn't make logical sense to handle object removal in this function, but it does make it very easy to guarantee
	// that we are for sure sending out at least one update in which the object is flagged as deleted
	for _, gameObj := range removed {
		delete(players.Players, gameObj.UUID) // thank you whoever made this  null safe
	}
	simulationObjects = removeIndexes(simulationObjects, removedIndexes)
	return &data
}

type bullet struct {
	*shared_structs.GameObject
	expirationDate time.Time
}

type tile struct {
	*shared_structs.GameObject
}

func (b bullet) ApplyBehavior(deltaTime float64) {
	if b.expirationDate.UnixMilli() < time.Now().UnixMilli() {
		b.GameObject.Delete = true
	}
}

func (t tile) ApplyBehavior(deltaTime float64) {
	return
}

func (b bullet) GetObject() *shared_structs.GameObject {
	return b.GameObject
}

func (t tile) GetObject() *shared_structs.GameObject {
	return t.GameObject
}

// removeIndexes removes elements from a slice at the given indices.
// The indices should be sorted in descending order to avoid issues with shifting elements.
func removeIndexes[T any](s []T, indexes []int) []T {
	// Sort indexes in descending order to handle shifting correctly.
	// This ensures that removing an element doesn't affect the indices
	// of elements that are yet to be removed.
	sort.Sort(sort.Reverse(sort.IntSlice(indexes)))

	for _, idx := range indexes {
		if idx >= 0 && idx < len(s) {
			s = append(s[:idx], s[idx+1:]...)
		}
	}
	return s
}
