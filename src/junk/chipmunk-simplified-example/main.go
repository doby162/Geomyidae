// https://github.com/oliverbestmann/box2d-go/tree/main/example
package main

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	cp "github.com/jakecoffman/cp/v2"
)

const Layers = 20

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(1200, 800)
	ops := ebiten.RunGameOptions{}
	ops.SingleThread = true
	_ = ebiten.RunGameWithOptions(&viz{}, &ops)
}

type viz struct {
	physics     Physics
	bodies      []Body
	physicsTime float64
	paused      bool
	slowmo      bool
	timeAcc     float64
}

func (v *viz) Update() error {
	// Initialize box2d physics on first update
	// I'm not sure if this could be done in some other place like an "init" function,
	// but this works for now.
	if v.physics == nil {
		v.initializePhysics()
	}

	// Handle input
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		v.paused = !v.paused
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		v.slowmo = !v.slowmo
	}

	// Step physics, otherwise it doesn't happen
	if v.physics != nil && !v.paused {
		var step = 1 / 60.0

		if v.slowmo {
			step *= 0.1
		}

		v.timeAcc += step

		for v.timeAcc > 1/60.0 {
			v.timeAcc -= 1 / 60.0

			startTime := time.Now()
			v.physics.Step(1/60.0, 4)

			if v.physicsTime == 0 {
				v.physicsTime = time.Since(startTime).Seconds()
			} else {
				v.physicsTime = 0.95*v.physicsTime + 0.05*time.Since(startTime).Seconds()
			}
		}
	}

	return nil
}

func (v *viz) Draw(screen *ebiten.Image) {
	// Remember, this just draws stuff, not update physics
	w, h := float64(screen.Bounds().Dx()), float64(screen.Bounds().Dy())

	var toScreen ebiten.GeoM

	var minX, minY, maxX, maxY float64
	for _, body := range v.bodies {
		x, y := body.Position()
		minX = min(minX, x)
		maxX = max(maxX, x)
		minY = min(minY, y)
		maxY = max(minY, y)
	}

	_ = maxY

	// scale := (h - 50) / (Layers + 5)
	scale := w / (5 * Layers)

	toScreen.Scale(scale, -scale)
	toScreen.Translate(w/2, h-50)

	v.physics.Draw(screen, toScreen)

	numGoRoutines := runtime.NumGoroutine()
	text := fmt.Sprintf("physics %T: %1.2fms - goroutines: %d\nPress 'p' to toggle pause.\nPress 's' to toggle slowmo.", v.physics, v.physicsTime*1000.0, numGoRoutines)
	ebitenutil.DebugPrintAt(screen, text, 475, 16)

	if v.paused {
		ebitenutil.DebugPrintAt(screen, "paused", 475, 64)
	}
}

func (v *viz) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}

func (v *viz) initializePhysics() {
	physics := cpNew(-10)

	var bodies []Body

	floor := BodyDef{Elasticity: 0.1, Friction: 0.9, Density: 1}
	box := BodyDef{Elasticity: 0.25, Friction: 0.5, Density: 1}
	big := BodyDef{Density: 10, Elasticity: 0.1, Friction: 0.9}

	bodies = append(bodies, physics.CreateStaticLine(-2*Layers, 0, 2*Layers, 0, floor))
	bodies = append(bodies, physics.CreateStaticLine(-2*Layers, 0, -2*Layers, 100, floor))
	bodies = append(bodies, physics.CreateStaticLine(2*Layers, 0, 2*Layers, 100, floor))

	// create layers of boxes
	for l := range Layers {
		for i := range l {
			centerX := float64(i) + 0.5 - float64(l)/2
			centerY := (Layers - float64(l) - 0.5) * 1.0

			bodies = append(bodies, physics.CreateSquare(0.5, centerX, centerY, box))
		}
	}

	// create a fast moving box
	b := physics.CreateSquare(2, -1.8*Layers, 2.6*Layers, big)
	b.SetVelocity(75, -100)
	bodies = append(bodies, b)

	v.physics = physics
	v.bodies = bodies
	v.paused = false
}

type Body interface {
	Position() (float64, float64)
	Rotation() float64
	SetVelocity(x, y float64)
}

type BodyDef struct {
	Density    float64
	Friction   float64
	Elasticity float64
}

type Physics interface {
	Draw(screen *ebiten.Image, toScreen ebiten.GeoM)
	Step(dt float64, subSteps int)
	CreateSquare(halfSize, centerX, centerY float64, def BodyDef) Body
	CreateStaticLine(x0, y0, x1, y1 float64, def BodyDef) Body
}

func cpNew(gravity float64) Physics {
	space := cp.NewSpace()
	space.SetGravity(cp.Vector{Y: gravity})
	return Chipmunk{Space: space}
}

type Chipmunk struct {
	Space *cp.Space
}

func (ph Chipmunk) Step(dt float64, subSteps int) {
	ph.Space.Iterations = uint(subSteps)
	ph.Space.Step(dt)
}

func (ph Chipmunk) Draw(screen *ebiten.Image, toScreen ebiten.GeoM) {
	cp.DrawSpace(ph.Space, drawer{screen, toScreen})
}

func (ph Chipmunk) CreateSquare(halfSize, centerX, centerY float64, d BodyDef) Body {
	body := cp.NewBody(1, 1)

	shape := cp.NewBox(body, halfSize*2, halfSize*2, 0)
	shape.SetElasticity(d.Elasticity)
	shape.SetDensity(d.Density)
	shape.SetFriction(d.Friction)
	body.AddShape(shape)
	body.SetPosition(cp.Vector{X: centerX, Y: centerY})

	ph.Space.AddShape(shape)
	ph.Space.AddBody(body)

	return &cpBody{
		body: body,
	}
}

func (ph Chipmunk) CreateStaticLine(x0, y0, x1, y1 float64, d BodyDef) Body {
	body := cp.NewStaticBody()

	a := cp.Vector{X: x0, Y: y0}
	b := cp.Vector{X: x1, Y: y1}
	shape := cp.NewSegment(body, a, b, 0)
	shape.SetElasticity(d.Elasticity)
	shape.SetDensity(d.Density)
	shape.SetFriction(d.Friction)
	body.AddShape(shape)

	ph.Space.AddShape(shape)
	ph.Space.AddBody(body)

	return &cpBody{
		body: body,
	}
}

type cpBody struct {
	body *cp.Body
}

func (b *cpBody) Position() (float64, float64) {
	pos := b.body.Position()
	return pos.X, pos.Y
}

func (b *cpBody) Rotation() float64 {
	rot := b.body.Rotation()
	return math.Atan2(rot.Y, rot.X)
}

func (b *cpBody) SetVelocity(x, y float64) {
	b.body.SetVelocity(x, y)
}

type drawer struct {
	Screen   *ebiten.Image
	toScreen ebiten.GeoM
}

func (d drawer) DrawCircle(pos cp.Vector, angle, radius float64, outline, fill cp.FColor, data any) {
	var p vector.Path

	x, y := d.toScreen.Apply(pos.X, pos.Y)
	radius, _ = d.toScreen.Apply(radius, 0)
	p.Arc(float32(x), float32(y), float32(radius), 0, 2*math.Pi, vector.Clockwise)

	dop := vector.DrawPathOptions{ColorScale: toColorScaleF(outline)}
	vector.StrokePath(d.Screen, &p, &vector.StrokeOptions{Width: 1}, &dop)

	dop.ColorScale = toColorScaleF(fill)
	vector.FillPath(d.Screen, &p, nil, &dop)
}

func (d drawer) DrawSegment(a, b cp.Vector, fill cp.FColor, data any) {
	var p vector.Path

	x, y := d.toScreen.Apply(a.X, a.Y)
	p.LineTo(float32(x), float32(y))

	x, y = d.toScreen.Apply(b.X, b.Y)
	p.LineTo(float32(x), float32(y))

	dop := vector.DrawPathOptions{ColorScale: toColorScaleF(fill)}
	vector.StrokePath(d.Screen, &p, &vector.StrokeOptions{Width: 1}, &dop)
}

func (d drawer) DrawFatSegment(a, b cp.Vector, radius float64, outline, fill cp.FColor, data any) {
	var p vector.Path

	x, y := d.toScreen.Apply(a.X, a.Y)
	p.LineTo(float32(x), float32(y))

	x, y = d.toScreen.Apply(b.X, b.Y)
	p.LineTo(float32(x), float32(y))

	// TODO make this actually a capsule based on radius

	dop := vector.DrawPathOptions{ColorScale: toColorScaleF(outline)}
	vector.StrokePath(d.Screen, &p, &vector.StrokeOptions{Width: 1}, &dop)

	dop.ColorScale = toColorScaleF(fill)
	vector.FillPath(d.Screen, &p, nil, &dop)
}

func (d drawer) DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill cp.FColor, data any) {
	var p vector.Path

	for _, v := range verts[:count] {
		x, y := d.toScreen.Apply(v.X, v.Y)
		p.LineTo(float32(x), float32(y))
	}
	p.Close()

	dop := vector.DrawPathOptions{ColorScale: toColorScaleF(outline)}
	vector.StrokePath(d.Screen, &p, &vector.StrokeOptions{Width: 1}, &dop)

	dop.ColorScale = toColorScaleF(fill)
	vector.FillPath(d.Screen, &p, nil, &dop)
}

func (d drawer) DrawDot(size float64, pos cp.Vector, fill cp.FColor, data any) {
	d.DrawCircle(pos, 2*math.Pi, 2, cp.FColor{}, fill, data)
}

func (d drawer) Flags() uint {
	return cp.DRAW_SHAPES
}

func (d drawer) OutlineColor() cp.FColor {
	return cp.FColor{R: 1, G: 1, B: 1, A: 1}
}

func (d drawer) ShapeColor(shape *cp.Shape, data any) cp.FColor {
	return cp.FColor{R: 1, G: 1, B: 1, A: 0.1}
}

func (d drawer) ConstraintColor() cp.FColor {
	return cp.FColor{R: 1, G: 0, B: 0, A: 1}
}

func (d drawer) CollisionPointColor() cp.FColor {
	return cp.FColor{}
}

func (d drawer) Data() any {
	return nil
}

func toColorScaleF(f cp.FColor) ebiten.ColorScale {
	var c ebiten.ColorScale
	c.Scale(f.R, f.G, f.B, 1)
	c.ScaleAlpha(f.A)
	return c
}
