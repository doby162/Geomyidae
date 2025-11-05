package sock_server

import (
	"fmt"
	"log"
	"net/http"

	"Geomyidae/server/player"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Api(playerList *player.List) *Hub {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	hub := newHub(playerList)
	go hub.run()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	portNumber := 8080
	listenServerURL := fmt.Sprintf(":%d", portNumber)
	log.Printf("Websocket server is now running on port %v \n", portNumber)
	go func() {
		err := http.ListenAndServe(listenServerURL, r)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}()
	return hub
}
