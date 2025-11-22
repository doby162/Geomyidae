# Geomyidae

a game

## Development

Development is possible on Windows, Linux or MacOS. If it isn't working for you, file an Issue.

### Go

This is written entirely in [Go](https://golang.org/), so start by installing Go on your system:  
https://go.dev/doc/install

### Ebitengine

The front end is built using [Ebitengine](https://ebitengine.org/), a Go based game engine.  
You don't need to DO anything to make Ebitengine work, the code setup will do that, but if you are having trouble and want to make sure that your environment isn't screwed up, follow the instructions here to validate your environment before preceding:  
https://ebitengine.org/en/documents/install.html#Confirming_your_environment

### Steps to set up your environment

Assuming you have Go installed, and maybe you ran the "Confirming your environment" command, the following should get you going:

1. Clone the repo.
2. `cd src`
3. `go run ./server/` will launch an instance of the server.
4. `go run ./client/` will launch an instance of the client.

### Hot reloading

You can use [air](https://github.com/air-verse/air) to have the client re-build and run whenever you make changes:  
One time, to install [air](https://github.com/air-verse/air), run:  
`go install github.com/air-verse/air@latest`

Then `cd src` and run `air` to start the client and have it re-build and re-run whenever you change code.

_Remember that you must also have the server running._

## Project structure

### front end
The front end is very small, code wise. It has two duties:
1. Display game state to the user
2. Relay the state of the user's keyboard to the server

## back end

### Object model
Every item in the game has to be rendered to the client, and fulfil whatever behaviors have been ascribed to it.
As such, each item is represented as a superset of the GameObject struct, and implements the HasBehavior interface,
The GameObject contains everything required to render to the client, and any state that needs to be generically accessed.
The main game loop doesn't know what kind of GameObject it's handling, so if state needs to be exposed, it is generally on the
GameObject. 

In addition to implementing game logic and containing display data, game objects contain references to the physics body associated with an object.
In order to better handle collision logic, the physics body also contains a reference to it's parent gameObject.

### main loop
the server's main function instantiates chipmunk physics, spawns in tiles to match the game map, and then runs the physics
and calls hasBehavior on objects. As such, there are two layers to the game. The physics layer, and the logic layer.

When possible, I prefer to let the physics layer handle things for me. So bullet knock back is implemented by making bullets heavy,
and the impact has knock back due to physics.
CCollisions are detected by chipmunks, but are handled in HasBehavior. Each GameObject has to handle it's own collision logic
if it wants something other than bounce.

## communication
As I have essentially re-invented OOP, objects in the simulation need to be able to talk to each other.
Here are some of the ways that happens:
1. each object's behavior method gives it access to the physics bodies it touches, which in turn own references to the associated game object.
2. each game object can contain an inbox, used to send strings between objects.
3. The main game loop has a spawnerPipeline. If an object creates a gameobject, it can be injected into the world via this channel
4. game objects can have flags. This is probably the best way to handle object deletes, but I'm trying to avoid the proliferation of flags

## best practices

The game is a work in progress and is a bit of a mess. But here are some code standards I am currently attempting to keep:
1. each game object should have an Identity that has a UserDataCode
2. each game object should instantiate and check an inbox if it has complex collision logic. Like an ammo pickup, it deletes itself and messages the player to increase its own ammo counter. This is preferable to both the player and the pickup each checking their own collisions.
3. when a game object creates a game object, it should do so via the spawnerPipeline
4. Each type of game object should have a constructor that handles as much of the logic as possible.
5. channels should not be sent to without a select and default. If a channel filled up and blocked, we would want to throw away extra data and maybe log an error rather than cause a deadlock
6. 