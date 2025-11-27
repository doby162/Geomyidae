package tile

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"Geomyidae/server/tracker"
	"Geomyidae/server/turret"
	"time"

	"github.com/jakecoffman/cp/v2"
)

type Tile struct {
	ActionSequence []Action
	*shared_structs.GameObject
}

type Action struct {
	Seconds int
	Type    constants.UserDataCode
	X       float64
	Y       float64
}

func NewTile(gameObject *shared_structs.GameObject, seq []Action) *Tile {
	return &Tile{seq, gameObject}
}

func (t *Tile) ApplyBehavior(deltaTime float64, spawnerPipeline chan shared_structs.HasBehavior) {
	t.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Player {
				if t.ActionSequence != nil {
					go func() {
						for _, action := range t.ActionSequence {
							time.Sleep(time.Duration(action.Seconds) * time.Second)
							var obj shared_structs.HasBehavior
							if action.Type == constants.Turret {
								obj = turret.NewTurret(ptr, action.X, action.Y)
							} else if action.Type == constants.Tracker {
								obj = tracker.NewTracker(ptr, action.X, action.Y)
							}
							if obj != nil {
								spawnerPipeline <- obj
							}
						}
					}()
					t.Delete = true
				}
			}
		}
	})
	return
}

func (t *Tile) GetObject() *shared_structs.GameObject {
	return t.GameObject
}
