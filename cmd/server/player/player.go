package player

import (
	"Geomyidae/cmd/server/box"
	"log"
	"math/rand"
	"sync"
	"time"
)

type List struct {
	Players     map[string]*NetworkPlayer
	physics     box.Physics
	writeAccess sync.Mutex
}

func NewList(physics box.Physics) *List {
	players := make(map[string]*NetworkPlayer)
	return &List{Players: players, writeAccess: sync.Mutex{}, physics: physics}
}

type NetworkPlayer struct {
	sprite   string
	canJump  bool
	name     string
	body     box.Body
	heldKeys []string
}

func (l *List) NewNetworkPlayer(sprite string) *NetworkPlayer {
	l.writeAccess.Lock()
	defer l.writeAccess.Unlock()
	name := generateRandomString(10)
	bd := box.BodyDef{Elasticity: 0.25, Friction: 0.0, Density: 1}

	body := l.physics.CreatePlayerCollider(0.5, 3, 3, bd, 1, 0.1)
	l.Players[name] = &NetworkPlayer{sprite: sprite, heldKeys: []string{}, name: name, canJump: true, body: body}
	return l.Players[name]
}

func (p *NetworkPlayer) ApplyKeys() {
	// todo
	log.Println(p.heldKeys)
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
