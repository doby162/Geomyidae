package main

import (
	"Geomyidae/cmd/server/sock_server"
)

func main() {
	err := sock_server.Api()
	if err != nil {
		return
	}
}
