package bullet

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"time"

	"github.com/jakecoffman/cp/v2"
)

type Bullet struct {
	*shared_structs.GameObject
	expirationDate time.Time
}

func NewBullet(gameObject *shared_structs.GameObject) *Bullet {
	return &Bullet{
		GameObject:     gameObject,
		expirationDate: time.Now().Add(time.Second * 5),
	}
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
