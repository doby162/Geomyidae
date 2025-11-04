package sock_server

import (
	"Geomyidae/server/player"
)

// Hub maintains the set of active Clients and broadcasts messages to the
// Clients.
type Hub struct {
	// Registered Clients.
	Clients map[*Client]bool

	// Inbound messages from the Clients.
	Broadcast chan []byte

	// Register requests from the Clients.
	register chan *Client

	// Unregister requests from Clients.
	unregister chan *Client

	playerList *player.List
}

func newHub(list *player.List) *Hub {
	return &Hub{
		playerList: list,
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			client.Player = h.playerList.NewNetworkPlayer()
			h.Clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				h.playerList.WriteAccess.Lock()
				h.playerList.Physics.RemoveShape(client.Player.Shape)
				h.playerList.Physics.RemoveBody(client.Player.Body)
				delete(h.playerList.Players, client.Player.Name)
				h.playerList.WriteAccess.Unlock()
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
