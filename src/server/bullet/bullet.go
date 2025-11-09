package bullet

import (
	"Geomyidae/internal/shared_structs"
	"time"
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

func (b Bullet) GetObject() *shared_structs.GameObject {
	return b.GameObject
}

func (b Bullet) ApplyBehavior(deltaTime float64) {
	if b.expirationDate.UnixMilli() < time.Now().UnixMilli() {
		b.GameObject.Delete = true
	}
}
