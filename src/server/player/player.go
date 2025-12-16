package player

import (
	"Geomyidae/internal/constants"
	"Geomyidae/internal/shared_structs"
	"Geomyidae/server/bomb"
	"Geomyidae/server/bullet"
	"math"
	"sync"

	"github.com/google/uuid"
	"github.com/jakecoffman/cp/v2"
)

type List struct {
	Players     map[string]*NetworkPlayer
	Physics     *cp.Space
	WriteAccess sync.Mutex
	apOb        func(networkPlayer *NetworkPlayer)
}

func NewList(physics *cp.Space, fn func(networkPlayer *NetworkPlayer)) *List {
	players := make(map[string]*NetworkPlayer)
	return &List{Players: players, WriteAccess: sync.Mutex{}, Physics: physics, apOb: fn}
}

type NetworkPlayer struct {
	*shared_structs.GameObject

	canJump              bool
	HeldKeys             []string
	shootTime            float64
	bombCount            int
	bombTime             float64
	portalToggleCooldown float64
}

// NewNetworkPlayer creates a network player and stores a pointer to it in both the master list and the network player list
// both values are the same pointer, it does not matter which you use, but you cannot reassign the pointer later
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

	l.Physics.AddShape(shape)
	l.Physics.AddBody(body)

	player := NetworkPlayer{GameObject: &shared_structs.GameObject{
		Sprite:        "spaceShooterRedux",
		UUID:          name,
		Body:          body,
		Shape:         shape,
		SpriteOffsetX: 325,
		SpriteOffsetY: 0,
		SpriteWidth:   98,
		SpriteHeight:  75,
		NeedsStatics:  true,
		Identity:      constants.Player,
		Inbox:         make(chan string, 10), // if the inbox fills up it will block so the sender is responsible for not sending data once it is full
		Portal:        true,
	}, HeldKeys: []string{}, canJump: true}
	body.UserData = player.GameObject
	point := &player
	l.apOb(point)
	l.Players[name] = point

	player.bombCount = 1
	return &player
}

// these could be multiplied by delta time
const thrust = 2
const maxSpeed = 25.0
const turn = 2

func (p *NetworkPlayer) ApplyBehavior(deltaTime float64, spawnerPipeline chan shared_structs.HasBehavior) {
	tr := thrust * deltaTime
	tn := turn * deltaTime
	x, y := p.Body.Velocity().X, p.Body.Velocity().Y
	if math.Abs(x)+math.Abs(y) > maxSpeed {
		p.Body.SetVelocityVector(p.Body.Velocity().Mult(0.95))
	}
	if p.shootTime >= 0 {
		p.shootTime -= deltaTime
	}
	if p.bombTime >= 0 {
		p.bombTime -= deltaTime
	}
	if p.portalToggleCooldown >= 0 {
		p.portalToggleCooldown -= deltaTime
	}
	select {
	case msg, ok := <-p.Inbox:
		if ok {
			if msg == "bombplus" {
				p.bombCount++
			}
		}
	default:
	}
	for _, key := range p.HeldKeys {
		if key == "B" && p.bombCount > 0 && p.bombTime <= 0 {
			p.bombCount--
			newBomb := bomb.NewBomb(float64(p.X), float64(p.Y))
			select {
			case spawnerPipeline <- newBomb:
			default:
			}
			p.bombTime = 0.5
		}
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
			newBullet := bullet.NewBullet(p.GetObject())
			select {
			case spawnerPipeline <- newBullet:
			default:
			}
		}
		if key == "P" && p.portalToggleCooldown <= 0 {
			p.Portal = !p.Portal
			p.portalToggleCooldown = 0.5
		}
	}

	return
}

func (p *NetworkPlayer) GetObject() *shared_structs.GameObject {
	return p.GameObject
}
