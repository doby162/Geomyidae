package tracker

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"math"

	"github.com/jakecoffman/cp/v2"
)

type Tracker struct {
	*shared_structs.GameObject
	target *shared_structs.GameObject
}

func NewTracker(gameObject, target *shared_structs.GameObject) *Tracker {
	return &Tracker{gameObject, target}
}

// even infinitesimal thrust gets fast quick
const thrust = 0.0001

func (t *Tracker) ApplyBehavior(deltaTime float64) {
	tr := thrust * deltaTime
	tpos := t.target.Body.Position()
	pos := t.Body.Position()
	angle := math.Atan2(tpos.Y-pos.Y, tpos.X-pos.X)
	t.Body.SetAngle(angle + (math.Pi / 2))
	t.Body.ApplyImpulseAtLocalPoint(cp.Vector{
		X: -math.Sin(tr),
		Y: -math.Cos(-tr),
	}, cp.Vector{X: 0, Y: 0})
	if t.target.Delete {
		t.Delete = true
	}
	t.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Player || ptr.Identity == constants.Bullet {
				t.Delete = true
			}
		}
	})
}

func (t *Tracker) GetObject() *shared_structs.GameObject {
	return t.GameObject
}
