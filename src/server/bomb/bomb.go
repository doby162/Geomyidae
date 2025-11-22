package bomb

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type Bomb struct {
	*shared_structs.GameObject
	expirationDate time.Time
	detCount       float64
	detMax         float64
	det            bool
}

func NewBomb(x, y float64) *Bomb {
	body := cp.NewBody(1, 1)
	shape := cp.NewBox(body, 1, 1, 0)
	shape.SetElasticity(0.25)
	shape.SetDensity(0.5)
	shape.SetFriction(1.0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: x / 64, Y: y / 64})
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
		Identity:      constants.Bomb,
	}
	body.UserData = &gameObject

	return &Bomb{
		GameObject:     &gameObject,
		expirationDate: time.Now().Add(time.Duration(1) * time.Second),
		detCount:       0,
		detMax:         36,
		det:            false,
	}
}
func (b *Bomb) ApplyBehavior(deltaTime float64) {
	if !b.det {
		if b.expirationDate.UnixMilli() < time.Now().UnixMilli() {
			b.det = true
		}
	}
	if b.detCount > b.detMax {
		b.Delete = true
	} else if b.det {
		degree := (math.Pi * 2) / b.detMax
		b.Body.SetAngle(degree * b.detCount)
		b.ShootFlag = true
		b.detCount++
	}
	b.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Bullet {
				b.Delete = true
			}
		}
	})
}

func (b *Bomb) GetObject() *shared_structs.GameObject {
	return b.GameObject
}
