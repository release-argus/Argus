// Copyright [2022] [Argus]
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

package v1

import (
	"encoding/json"
	"fmt"

	"github.com/release-argus/Argus/utils"
)

// Hub maintains the set of active clients and broadcasts messages to those
// clients.
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

// NewHub will create a new Hub.
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// AnnounceMSG is a minimal JSON struct to check whether the incoming message
// is valid.
type AnnounceMSG struct {
	Type      string `json:"type"`
	ServiceID string `json:"service_id"`
}

// Run will start the WebSocket Hub.
func (h *Hub) Run(jLog *utils.JLog) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			jLog.Debug(
				fmt.Sprintf("Broadcast %s", string(message)),
				utils.LogFrom{Primary: "WebSocket"},
				len(h.clients) > 0,
			)
			var msg AnnounceMSG
			if err := json.Unmarshal(message, &msg); err != nil {
				// TODO: Don't panic on invalid message?
				panic(err)
			}
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
