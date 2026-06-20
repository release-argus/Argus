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
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	Broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// AnnounceMSG is minimal JSON to validate the incoming message.
type AnnounceMSG struct {
	Type      string `json:"type"`
	ServiceID string `json:"service_id"`
}

// Run starts the WebSocket Hub.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Avoid unnecessary writes to the map.
			if _, ok := h.clients[client]; !ok {
				h.clients[client] = true
			}
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			if logx.IsLevel("DEBUG") {
				logx.Debug(
					"Broadcast "+string(message),
					logx.LogFrom{Primary: "WebSocket"},
					len(h.clients) > 0,
				)
			}

			// Validate JSON.
			var msg AnnounceMSG
			if err := decode.Unmarshal("json", message, &msg); err != nil {
				logx.Warn(
					"Invalid JSON broadcast to the WebSocket",
					logx.LogFrom{Primary: "WebSocket"},
					true,
				)
				continue
			}

			// Send message to all clients.
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
