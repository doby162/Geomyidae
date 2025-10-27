package sock_server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Api() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	hub := newHub()
	go hub.run()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	port := ":8080"
	log.Printf("Listening on http://127.0.0.1%v \n", port)
	return http.ListenAndServe(port, r)
	//return http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", r)
}
