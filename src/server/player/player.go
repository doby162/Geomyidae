package player

import (
	"math/rand"
	"sync"
	"time"

	"github.com/jakecoffman/cp/v2"
)

type List struct {
	Players     map[string]*NetworkPlayer
	Physics     *cp.Space
	WriteAccess sync.Mutex
}

func NewList(physics *cp.Space) *List {
	players := make(map[string]*NetworkPlayer)
	return &List{Players: players, WriteAccess: sync.Mutex{}, Physics: physics}
}

type NetworkPlayer struct {
	Sprite   string
	canJump  bool
	Name     string
	Body     *cp.Body
	Shape    *cp.Shape
	HeldKeys []string
}

func (l *List) NewNetworkPlayer() *NetworkPlayer {
	l.WriteAccess.Lock()
	defer l.WriteAccess.Unlock()
	name := generateRandomString(10)

	body := cp.NewBody(1, 1)
	shape := cp.NewBox(body, 1, 1, 0)
	shape.SetElasticity(0.25)
	shape.SetDensity(0.5)
	shape.SetFriction(1.0)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: 5, Y: 5})

	l.Physics.AddShape(shape)
	l.Physics.AddBody(body)

	l.Players[name] = &NetworkPlayer{Sprite: "player_01", HeldKeys: []string{}, Name: name, canJump: true, Body: body, Shape: shape}
	return l.Players[name]
}

var jump = 0.1

func (p *NetworkPlayer) ApplyKeys() {
	for _, key := range p.HeldKeys {
		if key == "W" {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: 0,
				Y: -jump,
			}, cp.Vector{X: 0, Y: 0})
		}
		if key == "A" {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: -jump,
				Y: 0,
			}, cp.Vector{X: 0, Y: 0})
		}
		if key == "D" {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: jump,
				Y: 0,
			}, cp.Vector{X: 0, Y: 0})
		}
		if key == "S" {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: 0,
				Y: jump,
			}, cp.Vector{X: 0, Y: 0})
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
