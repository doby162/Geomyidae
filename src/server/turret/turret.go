package turret

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"Geomyidae/server/bullet"
	"Geomyidae/server/pickup"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type Turret struct {
	*shared_structs.GameObject
	target    *shared_structs.GameObject
	shootTime time.Time
}

func (t *Turret) ApplyBehavior(deltaTime float64, spawnerPipeline chan shared_structs.HasBehavior) {
	tpos := t.target.Body.Position()
	pos := t.Body.Position()
	angle := math.Atan2(tpos.Y-pos.Y, tpos.X-pos.X)
	// the fact that I have to add half a pi of radians makes me very suspicious that we are converting angles incorrectly
	// probably in the client?
	t.Body.SetAngle(angle + (math.Pi / 2))
	if t.shootTime.UnixMilli() < time.Now().UnixMilli() {
		newBullet := bullet.NewBullet(t.GetObject())
		select {
		case spawnerPipeline <- newBullet:
		default:
		}
		t.shootTime = time.Now().Add(time.Second * 5)
	}
	if t.target.Delete {
		t.Delete = true
	}
	t.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if ptr, ok := bodB.UserData.(*shared_structs.GameObject); ok {
			if ptr.Identity == constants.Bullet {

				newPickup := pickup.NewPickup(pos.X, pos.Y, "bombplus")
				select {
				case spawnerPipeline <- newPickup:
				default:
				}
				t.Delete = true
			}
		}
	})
}

func (t *Turret) GetObject() *shared_structs.GameObject {
	return t.GameObject
}

func NewTurret(target *shared_structs.GameObject, x, y float64) *Turret {
	body := cp.NewBody(1, 1)
	shape := cp.NewBox(body, 1, 1, 0)
	shape.SetElasticity(0.25)
	shape.SetDensity(0.5)
	shape.SetFriction(1.0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: x, Y: y})
	obj := shared_structs.GameObject{
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
	}
	body.UserData = &obj
	return &Turret{&obj, target, time.Now().Add(time.Second * 5)}
}
