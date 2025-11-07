package player

import (
	"math"
	"sync"

	"github.com/google/uuid"
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
	Sprite       string
	canJump      bool
	Name         string
	Body         *cp.Body
	Shape        *cp.Shape
	HeldKeys     []string
	NeedsStatics bool
	shootTime    float64
	ShootFlag    bool
	Delete       bool
}

func (l *List) NewNetworkPlayer() *NetworkPlayer {
	l.WriteAccess.Lock()
	defer l.WriteAccess.Unlock()
	name := uuid.New().String()

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

	l.Players[name] = &NetworkPlayer{Sprite: "player_01", HeldKeys: []string{}, Name: name, canJump: true, Body: body,
		Shape: shape, NeedsStatics: true}
	return l.Players[name]
}

// these could be multiplied by delta time
const thrust = 2
const maxSpeed = 30.0
const turn = 2

func (p *NetworkPlayer) ApplyKeys(deltaTime float64) {
	//p.Body.EachArbiter(func(arbiter *cp.Arbiter) {
	//bodA, bodB := arbiter.Bodies()
	//})
	tr := thrust * deltaTime
	tn := turn * deltaTime
	x, y := p.Body.Velocity().X, p.Body.Velocity().Y
	if math.Abs(x)+math.Abs(y) > maxSpeed {
		p.Body.SetVelocityVector(p.Body.Velocity().Mult(0.95))
	}
	if p.shootTime >= 0 {
		p.shootTime -= deltaTime
	}
	for _, key := range p.HeldKeys {
		if key == "W" {
			p.Body.ApplyImpulseAtLocalPoint(cp.Vector{
				X: -math.Sin(tr),
				Y: -math.Cos(-tr),
			}, cp.Vector{X: 0, Y: 0})
		}
		if key == "A" {
			rot := p.Body.Angle()
			p.Body.SetAngle(rot - tn)
			p.Body.SetAngularVelocity(0)
		}
		if key == "D" {
			rot := p.Body.Angle()
			p.Body.SetAngle(rot + tn)
			p.Body.SetAngularVelocity(0)
		}
		if key == "S" {
			p.Body.SetVelocityVector(p.Body.Velocity().Mult(0.95))
			p.Body.SetAngularVelocity(p.Body.AngularVelocity() * 0.75)
		}
		if key == "E" && p.shootTime <= 0 {
			p.shootTime = 0.5 // seconds
			p.ShootFlag = true
		}
	}

	return
}
