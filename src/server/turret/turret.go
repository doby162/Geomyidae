package turret

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"math"
	"time"

	"github.com/jakecoffman/cp/v2"
)

type Turret struct {
	*shared_structs.GameObject
	target    *shared_structs.GameObject
	shootTime time.Time
}

func (t *Turret) ApplyBehavior(deltaTime float64) {
	tpos := t.target.Body.Position()
	pos := t.Body.Position()
	angle := math.Atan2(tpos.Y-pos.Y, tpos.X-pos.X)
	// the fact that I have to add half a pi of radians makes me very suspicious that we are converting angles incorrectly
	// probably in the client?
	t.Body.SetAngle(angle + (math.Pi / 2))
	if t.shootTime.UnixMilli() < time.Now().UnixMilli() {
		t.GameObject.ShootFlag = true
		t.shootTime = time.Now().Add(time.Second * 5)
	}
	if t.target.Delete {
		t.Delete = true
	}
	t.Body.EachArbiter(func(arbiter *cp.Arbiter) {
		_, bodB := arbiter.Bodies()
		if bodB.UserData == constants.Bullet {
			// we've been shot by something claiming to be a bullet!
			t.Delete = true
		}
	})
}

func (t *Turret) GetObject() *shared_structs.GameObject {
	return t.GameObject
}

func NewTurret(gameObject, target *shared_structs.GameObject) *Turret {
	return &Turret{gameObject, target, time.Now().Add(time.Second * 5)}
}
