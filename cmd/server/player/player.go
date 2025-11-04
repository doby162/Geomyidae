package player

import (
	"Geomyidae/cmd/server/box"
	"math/rand"
	"sync"
	"time"

	b2 "github.com/oliverbestmann/box2d-go"
)

type List struct {
	Players     map[string]*NetworkPlayer
	physics     box.Physics
	WriteAccess sync.Mutex
}

func NewList(physics box.Physics) *List {
	players := make(map[string]*NetworkPlayer)
	return &List{Players: players, WriteAccess: sync.Mutex{}, physics: physics}
}

type NetworkPlayer struct {
	Sprite   string
	canJump  bool
	Name     string
	Body     box.Body
	HeldKeys []string
}

func (l *List) NewNetworkPlayer() *NetworkPlayer {
	l.WriteAccess.Lock()
	defer l.WriteAccess.Unlock()
	name := generateRandomString(10)
	bd := box.BodyDef{Elasticity: 0.25, Friction: 0.0, Density: 1}

	body := l.physics.CreatePlayerCollider(0.5, 3, 3, bd, 1, 0.1)
	l.Players[name] = &NetworkPlayer{Sprite: "player_01", HeldKeys: []string{}, Name: name, canJump: true, Body: body}
	return l.Players[name]
}

var jump = float32(1.0)

func (p *NetworkPlayer) ApplyKeys() {
	for _, key := range p.HeldKeys {
		if key == "W" {
			p.Body.ApplyImpulse(b2.Vec2{
				X: 0,
				Y: -jump,
			})
		}
		if key == "A" {
			p.Body.ApplyImpulse(b2.Vec2{
				X: -jump,
				Y: 0,
			})
		}
		if key == "D" {
			p.Body.ApplyImpulse(b2.Vec2{
				X: jump,
				Y: 0,
			})
		}
	}

	return
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
