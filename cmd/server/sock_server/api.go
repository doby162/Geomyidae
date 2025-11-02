package sock_server

import (
	"log"
	"net/http"

	"Geomyidae/cmd/server/player"

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
	port := ":8080"
	log.Printf("Listening on http://127.0.0.1%v \n", port)
	go func() {
		err := http.ListenAndServe(port, r)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}()
	//return http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", r)
	return hub
}
