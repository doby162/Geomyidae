package main

import (
	"Geomyidae/cmd/server/box"
	"Geomyidae/cmd/server/player"
	"Geomyidae/cmd/server/sock_server"
	"log"
	"time"
)

// for this refactor I need to take all the physics logic and put it into the server
// the front end will use ebiten and the backend will use box2d
// socket messages will no longer be braodcast, rather they will set values that will be referenced
// by the game loop on the server
// regularly scheduled updates will be sent out to clients
// this is basically a thin client pattern where the client consists of a screen and a gamepad
// but the game fully takes place on the server

var physics box.Physics
var prevTime time.Time

func main() {
	playerList := player.NewList(physics)
	go func() {
		err := sock_server.Api(playerList)
		if err != nil {
			log.Fatal(err)
		}
	}()
	prevTime = time.Now()

	physics = box.B2New(9.8)

	for {
		deltaTime := time.Now().Sub(prevTime).Seconds()
		prevTime = time.Now()
		physics.Step(deltaTime, 1)
		for _, networkPlayer := range playerList.Players {
			networkPlayer.ApplyKeys()
		}
	}
}
