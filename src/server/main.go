package main

import (
	"Geomyidae/internal/constants"
	"Geomyidae/server/bomb"
	"Geomyidae/server/bullet"
	"Geomyidae/server/pickup"
	"Geomyidae/server/player"
	"Geomyidae/server/sock_server"
	"Geomyidae/server/tile"
	"Geomyidae/server/tracker"
	"Geomyidae/server/turret"
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

// simulationObjects may not look like a slice of pointers, but it is
// because HasBehavior is implemented with pointer receiver methods

// simulationObjects is not stored as a pointer because I append to it locally
// players is however, because it is updated remotely by the socket server
// apOb allows players to add new NetworkPlayers to simulationObjects remotely
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

	spawnerPipeline := make(chan shared_structs.HasBehavior, 10)

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

		obj := tile.NewTile(
			&shared_structs.GameObject{
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
				Identity:             constants.Tile,
			})
		body.UserData = obj.GameObject

		physics.AddShape(shape)
		simulationObjects = append(simulationObjects, obj)
		physics.AddBody(body)
	}

	newPickup := pickup.NewPickup(7, 7, "bombplus")
	physics.AddShape(newPickup.Shape)
	physics.AddBody(newPickup.Body)
	simulationObjects = append(simulationObjects, newPickup)

	// kick off socket server
	hub := sock_server.Api(players)

	prevTime = time.Now()
	turretSpawnTIme := time.Now()
	for {
		includeStaticAndAsleep := false
		players.WriteAccess.Lock()
		deltaTime := time.Now().Sub(prevTime).Seconds()
		countTurrets := 0

		for _, obj := range simulationObjects {
			obj.ApplyBehavior(deltaTime, spawnerPipeline)
			gameObj := obj.GetObject()
			if gameObj.NeedsStatics {
				log.Println("full packet")
				includeStaticAndAsleep = true
				gameObj.NeedsStatics = false
			}
			_, ok := obj.(*turret.Turret)
			if ok {
				countTurrets++
			}
			if gameObj.BombDrop {
				gameObj.BombDrop = false
				newBomb := bomb.NewBomb(gameObj.X, gameObj.Y)
				physics.AddBody(newBomb.Body)
				physics.AddShape(newBomb.Shape)
				simulationObjects = append(
					simulationObjects, newBomb)
			}
			if gameObj.ShootFlag {
				gameObj.ShootFlag = false
				body := cp.NewBody(1, 1)
				shape := cp.NewCircle(body, 0.125, cp.Vector{X: 0, Y: 0})
				shape.SetElasticity(0.25)
				shape.SetDensity(50.5)
				shape.SetFriction(1.0)
				body.AddShape(shape)
				pos := gameObj.Body.Position()
				x := pos.X
				y := pos.Y
				angle := gameObj.Body.Angle()
				thrust := 35.0
				offset := 1.0
				x = x + math.Sin(angle)*offset
				y = y + math.Cos(angle)*(offset*-1)
				body.SetVelocity(math.Sin(angle)*thrust, math.Cos(angle)*(-1*thrust))
				body.SetPosition(cp.Vector{X: x, Y: y})

				physics.AddBody(body)
				physics.AddShape(shape)
				newBullet := bullet.NewBullet(&shared_structs.GameObject{
					Sprite:        "spaceShooterRedux",
					SpriteOffsetX: 0,
					SpriteOffsetY: 0,
					SpriteWidth:   16,
					SpriteHeight:  16,
					Angle:         body.Angle(),
					UUID:          uuid.New().String(),
					Body:          body,
					Shape:         shape,
					Identity:      constants.Bullet,
				})

				simulationObjects = append(simulationObjects, newBullet)
				body.UserData = newBullet.GameObject
			}
		}

		select {
		case msg, ok := <-spawnerPipeline:
			if ok {
				obj := msg.GetObject()
				physics.AddBody(obj.Body)
				physics.AddShape(obj.Shape)
				simulationObjects = append(simulationObjects, msg)
			}
		default:
		}

		var targetBod *shared_structs.GameObject
		for _, pl := range players.Players {
			targetBod = pl.GameObject
		}

		if countTurrets < len(players.Players) && time.Now().UnixMilli() > turretSpawnTIme.Add(1*time.Second).UnixMilli() {
			turretSpawnTIme = time.Now()
			body := cp.NewBody(1, 1)
			shape := cp.NewBox(body, 1, 1, 0)
			shape.SetElasticity(0.25)
			shape.SetDensity(0.5)
			shape.SetFriction(1.0)
			body.AddShape(shape)
			body.SetPosition(cp.Vector{X: 5, Y: 5})
			physics.AddBody(body)
			physics.AddShape(shape)
			newTurret := turret.NewTurret(&shared_structs.GameObject{
				Sprite:               "spaceShooterRedux",
				SpriteOffsetX:        225,
				SpriteOffsetY:        0,
				SpriteWidth:          98,
				SpriteHeight:         75,
				SpriteFlipHorizontal: false,
				SpriteFlipVertical:   true,
				SpriteFlipDiagonal:   false,
				Angle:                0,
				UUID:                 uuid.New().String(),
				Body:                 body,
				Shape:                shape,
				Identity:             constants.Turret,
			}, targetBod)
			simulationObjects = append(simulationObjects, newTurret)
			body.UserData = newTurret.GameObject

			body = cp.NewBody(1, 1)
			shape = cp.NewBox(body, 1, 1, 0)
			shape.SetElasticity(0.25)
			shape.SetDensity(0.5)
			shape.SetFriction(1.0)
			body.AddShape(shape)
			body.SetPosition(cp.Vector{X: 5, Y: 5})
			physics.AddBody(body)
			physics.AddShape(shape)
			newTracker := tracker.NewTracker(&shared_structs.GameObject{
				Sprite:               "spaceShooterRedux",
				SpriteOffsetX:        450,
				SpriteOffsetY:        0,
				SpriteWidth:          98,
				SpriteHeight:         75,
				SpriteFlipHorizontal: false,
				SpriteFlipVertical:   true,
				SpriteFlipDiagonal:   false,
				Angle:                0,
				UUID:                 uuid.New().String(),
				Body:                 body,
				Shape:                shape,
				Identity:             constants.Tracker,
			}, targetBod)
			simulationObjects = append(simulationObjects, newTracker)
			body.UserData = newTracker.GameObject
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
		pruneWorldState()

		for sock, _ := range hub.Clients {
			data.Name = sock.Player.UUID
			msg, _ := json.Marshal(data)
			sock.Send <- msg
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func pruneWorldState() {
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
	}

	players.WriteAccess.Lock()
	defer players.WriteAccess.Unlock()
	for _, gameObj := range removed {
		delete(players.Players, gameObj.UUID)
	}

	simulationObjects = removeIndexes(simulationObjects, removedIndexes)
}

func collectWorldState(includeStaticAndAsleep bool) *shared_structs.WorldData {
	data := shared_structs.WorldData{}
	for _, obj := range simulationObjects {
		gameObj := obj.GetObject()
		if !includeStaticAndAsleep && gameObj.IsStatic {
			continue
		}
		if !includeStaticAndAsleep && gameObj.Body.IsSleeping() {
			continue
		}
		data.Objects = append(data.Objects, *obj.GetObject())
	}
	return &data
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
