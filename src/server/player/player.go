package player

import (
	"math"
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
	Sprite      string
	canJump     bool
	Name        string
	Body        *cp.Body
	Shape       *cp.Shape
	HeldKeys    []string
	thrustTimer int
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
	body.Rotation()

	l.Physics.AddShape(shape)
	l.Physics.AddBody(body)

	l.Players[name] = &NetworkPlayer{Sprite: "player_01", HeldKeys: []string{}, Name: name, canJump: true, Body: body, Shape: shape}
	return l.Players[name]
}

// these could be multiplied by delta time
var thrust = 0.02
var maxSpeed = 2.0
var turn = 0.02

func (p *NetworkPlayer) ApplyKeys() {
	for _, key := range p.HeldKeys {
		p.Body.EachArbiter(func(arbiter *cp.Arbiter) {
			// this does not work as well as I would like it to, currently
			// getting the actual collision data and doing something with it
			// is probably the key
			//bodA, bodB := arbiter.Bodies()
			p.thrustTimer = 200
		})
		if p.thrustTimer > 0 {
			p.thrustTimer -= 1
		}
		if key == "W" && p.thrustTimer == 0 {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: -math.Sin(thrust),
				Y: -math.Cos(-thrust),
			}, cp.Vector{X: 0, Y: 0})
		}
		if key == "A" {
			rot := p.Body.Angle()
			p.Body.SetAngle(rot - turn)
			p.Body.SetAngularVelocity(0)
		}
		if key == "D" {
			rot := p.Body.Angle()
			p.Body.SetAngle(rot + turn)
			p.Body.SetAngularVelocity(0)
		}
		if key == "S" {
			p.Body.SetVelocityVector(p.Body.Velocity().Mult(0.95))
			p.Body.SetAngularVelocity(0)
		}
		x, y := p.Body.Velocity().X, p.Body.Velocity().Y
		if math.Abs(x)+math.Abs(y) > maxSpeed {
			p.Body.SetVelocityVector(p.Body.Velocity().Mult(0.95))
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
