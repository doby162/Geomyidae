package tracker

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"math"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type Tracker struct {
	*shared_structs.GameObject
	target *shared_structs.GameObject
}

func NewTracker(target *shared_structs.GameObject, x, y float64) *Tracker {
	body := cp.NewBody(1, 1)
	shape := cp.NewBox(body, 1, 1, 0)
	shape.SetElasticity(0.25)
	shape.SetDensity(0.5)
	shape.SetFriction(1.0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: x, Y: y})
	obj := shared_structs.GameObject{
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
	}
	body.UserData = &obj

	return &Tracker{&obj, target}
}

// even infinitesimal thrust gets fast quick
const thrust = 0.0001

func (t *Tracker) ApplyBehavior(deltaTime float64, spawnerPipeline chan shared_structs.HasBehavior) {
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
