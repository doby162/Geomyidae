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
