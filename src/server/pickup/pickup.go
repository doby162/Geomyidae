package pickup

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type Pickup struct {
	*shared_structs.GameObject
	pickupType string
}

func (p *Pickup) ApplyBehavior(deltaTime float64) {
	p.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Player {
				p.Delete = true
				// if the channel is initialized and can accept a value currently, send bombplus
				if ptr.Inbox != nil {
					select {
					case ptr.Inbox <- p.pickupType:
					default:
					}
				}
			}
		}
	})
}

func NewPickup(x float64, y float64, str string) *Pickup {
	body := cp.NewBody(1, 1)
	shape := cp.NewBox(body, 1, 1, 0)
	shape.SetElasticity(0.25)
	shape.SetDensity(0.5)
	shape.SetFriction(1.0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: x, Y: y})
	gameObject := shared_structs.GameObject{
		X:             x,
		Y:             y,
		Sprite:        "spaceShooterRedux",
		SpriteOffsetX: 0,
		SpriteOffsetY: 0,
		SpriteWidth:   16,
		SpriteHeight:  16,
		UUID:          uuid.New().String(),
		Body:          body,
		Shape:         shape,
		IsStatic:      false,
		Identity:      constants.Pickup,
	}
	body.UserData = &gameObject

	return &Pickup{
		pickupType: str,
		GameObject: &gameObject,
	}
}

func (p *Pickup) GetObject() *shared_structs.GameObject {
	return p.GameObject
}
