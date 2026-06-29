// Copyright [2026] [Argus]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package v1 provides the API for the webserver.
package v1

import (
	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
)

// Hub maintains the set of active clients and broadcasts messages to those clients.
type Hub struct {
	// Registered clients (owned by the Run goroutine).
	clients map[*Client]bool

	Broadcast  chan []byte                 // Inbound messages from the clients.
	register   chan *Client                // Register requests from the clients.
	unregister chan *Client                // Unregister requests from clients.
	query      chan func(map[*Client]bool) // Observe clients on the Run goroutine.
}

// clientCount returns the number of registered clients.
func (h *Hub) clientCount() int {
	return len(h.clients)
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		query:      make(chan func(map[*Client]bool)),
		clients:    make(map[*Client]bool),
	}
}

// AnnounceMSG is minimal JSON to validate the incoming message.
type AnnounceMSG struct {
	Type      string `json:"type"`
	ServiceID string `json:"service_id"`
}

// addClient registers client.
func (h *Hub) addClient(client *Client) {
	if _, ok := h.clients[client]; !ok {
		h.clients[client] = true
	}
}

// removeClient unregisters client and closes its send channel.
func (h *Hub) removeClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

// broadcast sends message to every client, dropping any whose buffer is full.
func (h *Hub) broadcast(message []byte) {
	if logx.IsLevel("DEBUG") {
		logx.Debug(
			"Broadcast "+string(message),
			logx.LogFrom{Primary: "WebSocket"},
			h.clientCount() > 0,
		)
	}

	var msg AnnounceMSG
	if err := decode.Unmarshal("json", message, &msg); err != nil {
		logx.Warn(
			"Invalid JSON broadcast to the WebSocket",
			logx.LogFrom{Primary: "WebSocket"},
			true,
		)
		return
	}

	// Non-blocking send; drop any client whose buffer is full.
	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// Run starts the Hub. It owns the clients map; all access happens on this goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)
		case client := <-h.unregister:
			h.removeClient(client)
		case message := <-h.Broadcast:
			h.broadcast(message)
		case fn := <-h.query:
			fn(h.clients) // Read the map on the owning goroutine.
		}
	}
}
