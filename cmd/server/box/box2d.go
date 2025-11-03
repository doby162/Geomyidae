package box

import (
	"math"

	b2 "github.com/oliverbestmann/box2d-go"
)

type Body interface {
	ApplyForce(b2.Vec2)
	ApplyImpulse(b2.Vec2)
	Position() (float64, float64)
	Rotation() float64
	SetVelocity(x, y float64)
	SetPosition(x, y float64)
	DestroyBody()
}

type BodyDef struct {
	Density    float64
	Friction   float64
	Elasticity float64
}

type Physics interface {
	Step(dt float64, subSteps int)
	CreatePlayerCollider(halfSize, centerX, centerY float64, def BodyDef, fixedRotation uint8, damp float32) Body
	CreateStaticLine(x0, y0, x1, y1 float64, def BodyDef) Body
	CreateStaticTile(halfSize, centerX, centerY float64, def BodyDef) Body
	CreateNetworkCollider(halfSize, centerX, centerY float64, def BodyDef, fixedRotation uint8, damp float32) Body
}

func B2New(gravity float64) Physics {
	def := b2.DefaultWorldDef()
	def.Gravity = b2.Vec2{Y: float32(gravity)}

	b2.EnableConcurrency(&def)

	return &Box2D{
		World: b2.CreateWorld(def),
	}
}

type Box2D struct {
	World b2.World
}

func (ph Box2D) Step(dt float64, subSteps int) {
	ph.World.Step(float32(dt), int32(subSteps))
}

func (ph Box2D) CreatePlayerCollider(halfSize, centerX, centerY float64, d BodyDef, fixedRotation uint8, damp float32) Body {
	var tr b2.Transform
	tr.P.X = float32(centerX)
	tr.P.Y = float32(centerY)
	tr.Q.C = 1

	def := b2.DefaultBodyDef()
	def.LinearDamping = damp
	def.Type1 = b2.DynamicBody
	body := ph.World.CreateBody(def)
	body.SetTransform(tr.P, tr.Q)
	body.SetFixedRotation(fixedRotation)
	shape := b2.DefaultShapeDef()
	shape.Density = float32(d.Density)
	shape.Material.Restitution = float32(d.Elasticity)
	shape.Material.Friction = float32(d.Friction)
	body.CreateCapsuleShape(shape, b2.Capsule{
		Center1: b2.Vec2{X: float32(0), Y: float32(0)},
		Center2: b2.Vec2{X: float32(0), Y: float32(0)},
		Radius:  float32(halfSize),
	})

	return &b2Body{body: body}
}

func (ph Box2D) CreateNetworkCollider(halfSize, centerX, centerY float64, d BodyDef, fixedRotation uint8, damp float32) Body {
	var tr b2.Transform
	tr.P.X = float32(centerX)
	tr.P.Y = float32(centerY)
	tr.Q.C = 1

	def := b2.DefaultBodyDef()
	def.LinearDamping = damp
	def.Type1 = b2.StaticBody
	body := ph.World.CreateBody(def)
	body.SetTransform(tr.P, tr.Q)
	body.SetFixedRotation(fixedRotation)
	shape := b2.DefaultShapeDef()
	shape.Density = float32(d.Density)
	shape.Material.Restitution = float32(d.Elasticity)
	shape.Material.Friction = float32(d.Friction)
	body.CreateCapsuleShape(shape, b2.Capsule{
		Center1: b2.Vec2{X: float32(0), Y: float32(0)},
		Center2: b2.Vec2{X: float32(0), Y: float32(0)},
		Radius:  float32(halfSize),
	})

	return &b2Body{body: body}
}

func (ph Box2D) CreateStaticTile(halfSize, centerX, centerY float64, d BodyDef) Body {
	var tr b2.Transform
	tr.P.X = float32(centerX)
	tr.P.Y = float32(centerY)
	tr.Q.C = 1

	def := b2.DefaultBodyDef()
	def.Type1 = b2.StaticBody

	body := ph.World.CreateBody(def)
	body.SetTransform(tr.P, tr.Q)

	shape := b2.DefaultShapeDef()
	shape.Density = float32(d.Density)
	shape.Material.Restitution = float32(d.Elasticity)
	shape.Material.Friction = float32(d.Friction)
	body.CreatePolygonShape(shape, b2.MakeSquare(float32(halfSize)))

	return &b2Body{body: body}
}

func (ph Box2D) CreateStaticLine(x0, y0, x1, y1 float64, d BodyDef) Body {

	def := b2.DefaultBodyDef()
	def.Type1 = b2.StaticBody

	shape := b2.DefaultShapeDef()
	shape.Density = float32(d.Density)
	shape.Material.Restitution = float32(d.Elasticity)
	shape.Material.Friction = float32(d.Friction)

	body := ph.World.CreateBody(def)
	body.CreateSegmentShape(shape, b2.Segment{
		Point1: b2.Vec2{X: float32(x0), Y: float32(y0)},
		Point2: b2.Vec2{X: float32(x1), Y: float32(y1)},
	})

	return &b2Body{
		body: body,
	}
}

type b2Body struct {
	body b2.Body
}

func (b *b2Body) Position() (float64, float64) {
	pos := b.body.GetPosition()
	return float64(pos.X), float64(pos.Y)
}

func (b *b2Body) Rotation() float64 {
	rot := b.body.GetRotation()
	return math.Atan2(float64(rot.S), float64(rot.C))
}

func (b *b2Body) SetVelocity(x, y float64) {
	b.body.SetLinearVelocity(b2.Vec2{X: float32(x), Y: float32(y)})
}
func (b *b2Body) SetPosition(x, y float64) {
	v := b2.Vec2{X: float32(x), Y: float32(y)}
	b.body.SetTransform(v, b.body.GetRotation())
}

func (b *b2Body) ApplyForce(force b2.Vec2) {
	b.body.ApplyForce(force, b.body.GetPosition(), 1)
}
func (b *b2Body) ApplyImpulse(force b2.Vec2) {
	b.body.ApplyLinearImpulse(force, b.body.GetPosition(), 1)
}
func (b *b2Body) DestroyBody() {
	b.body.DestroyBody()
}
