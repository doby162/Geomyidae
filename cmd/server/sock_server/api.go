package sock_server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func Api() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	hub := newHub()
	go hub.run()
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	log.Println("Listening on :8000")
	return http.ListenAndServe(":8080", r)
	//return http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", r)
}
