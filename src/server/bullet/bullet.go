package bullet

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type Bullet struct {
	*shared_structs.GameObject
	expirationDate time.Time
}

func NewBullet(gameObj *shared_structs.GameObject) *Bullet {
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
	newBullet := Bullet{GameObject: &shared_structs.GameObject{
		Sprite:        "spaceShooterRedux",
		SpriteOffsetX: 0,
		SpriteOffsetY: 0,
		SpriteWidth:   16,
		SpriteHeight:  16,
		Angle:         shared_structs.RoundedFloat2(body.Angle()),
		UUID:          uuid.New().String(),
		Body:          body,
		Shape:         shape,
		Identity:      constants.Bullet,
	},
		expirationDate: time.Now().Add(time.Second * 5)}
	body.UserData = newBullet.GameObject

	return &newBullet
}

func (b *Bullet) GetObject() *shared_structs.GameObject {
	return b.GameObject
}

func (b *Bullet) ApplyBehavior(deltaTime float64, spawnerPipeline chan shared_structs.HasBehavior) {
	if b.expirationDate.UnixMilli() < time.Now().UnixMilli() {
		b.GameObject.Delete = true
	}
	b.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Player || ptr.Identity == constants.Turret || ptr.Identity == constants.Bullet {
				b.Delete = true
			}
		}
	})
}
