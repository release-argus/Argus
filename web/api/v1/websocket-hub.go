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

// AnnounceMSG is minimal JSON to validate the incoming message and identify
// which service (if any) it concerns.
type AnnounceMSG struct {
	Type        string `json:"type"`
	SubType     string `json:"sub_type"`
	ServiceID   string `json:"service_id"`
	ServiceData *struct {
		ID string `json:"id"`
	} `json:"service_data"`
}

// serviceID returns the ID of the service the message concerns
// ("" for messages not tied to a single service, e.g. the full ordering).
func (m *AnnounceMSG) serviceID() string {
	if m.ServiceData != nil && m.ServiceData.ID != "" {
		return m.ServiceData.ID
	}
	// DELETE and EDIT messages carry the service ID as their sub-type.
	if m.Type == "DELETE" || m.Type == "EDIT" {
		return m.SubType
	}
	return m.ServiceID
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
	serviceID := msg.serviceID()

	// Non-blocking send; drop any client whose buffer is full.
	for client := range h.clients {
		// Only send messages the client is permitted to see.
		if !client.mayReceive(serviceID) {
			continue
		}
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// kickMatching disconnects every client matching match. Kicked clients
// auto-reconnect, re-deriving their permitted-service sets.
func (h *Hub) kickMatching(match func(*Client) bool) {
	h.query <- func(clients map[*Client]bool) {
		// Runs on the Hub goroutine, so mutating the map is safe.
		for client := range clients {
			if match(client) {
				delete(clients, client)
				close(client.send)
			}
		}
	}
}

// KickUserClients disconnects the clients of userIDs. Call after anything
// that may change those users' grants (membership edits, group grant changes,
// disable/delete). Clients with no user identity (auth disabled) never match.
func (h *Hub) KickUserClients(userIDs ...string) {
	ids := make(map[string]bool, len(userIDs))
	for _, id := range userIDs {
		if id != "" {
			ids[id] = true
		}
	}
	if len(ids) == 0 {
		return
	}
	h.kickMatching(func(c *Client) bool { return ids[c.userID] })
}

// KickSessionClients disconnects the clients connected under the sessions
// with tokenHashes. Call after revoking those sessions so a revoked session
// can't keep an open WebSocket.
func (h *Hub) KickSessionClients(tokenHashes ...string) {
	hashes := make(map[string]bool, len(tokenHashes))
	for _, hash := range tokenHashes {
		if hash != "" {
			hashes[hash] = true
		}
	}
	if len(hashes) == 0 {
		return
	}
	h.kickMatching(func(c *Client) bool { return hashes[c.sessionHash] })
}

// KickRestrictedClients disconnects every client with a restricted
// permitted-service set. Call after service ID/tag changes, which change
// what those sets should match; unrestricted clients are unaffected.
func (h *Hub) KickRestrictedClients() {
	h.kickMatching(func(c *Client) bool { return c.allowedServices != nil })
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
