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
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/logx"
	apitype "github.com/release-argus/Argus/web/api/types"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// pingPeriod is the interval between WebSocket ping frames. Must occur before pongWait.
var pingPeriod = (pongWait * 9) / 10

// ServeWs upgrades an HTTP connection to WebSocket and registers the client with the hub.
// allowedServices limits which services' broadcasts the client receives (nil = unrestricted).
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, allowedServices map[string]bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	conn.RemoteAddr()
	client := &Client{
		hub:             hub,
		ip:              getIP(r),
		conn:            conn,
		send:            make(chan []byte, 256),
		allowedServices: allowedServices,
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all connections.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Client is a WebSocket connection registered with the Hub.
type Client struct {
	// The WebSocket hub.
	hub *Hub

	// The client's IP.
	ip string

	// The WebSocket connection.
	conn *websocket.Conn

	// send carries outbound messages from the hub/server to this client.
	send chan []byte

	// allowedServices limits which services' broadcasts this client receives
	// (nil = unrestricted). Derived from the user's permission grants at handshake.
	allowedServices map[string]bool
}

// mayReceive reports whether the client may receive a broadcast about serviceID
// ("" = a message not tied to a single service, e.g. the full service ordering
// - restricted clients never receive those).
func (c *Client) mayReceive(serviceID string) bool {
	if c.allowedServices == nil {
		return true
	}
	return serviceID != "" && c.allowedServices[serviceID]
}

// getIP returns the client IP resolved by [API.clientIPMiddleware],
// falling back to RemoteAddr when the middleware has not run.
func getIP(r *http.Request) string {
	if ip, ok := r.Context().Value(clientIPKey{}).(string); ok {
		return ip
	}
	return remoteAddrIP(r)
}

// forwardedIP returns the real client IP claimed by proxy headers ("" if none).
// Only call for requests from trusted proxies - the headers are spoofable.
//
// X-Forwarded-For is read right-to-left, returning the first address that is
// not itself a trusted proxy. CF-Connecting-Ip / X-Real-Ip are used as a
// fallback.
func (api *API) forwardedIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			addr, err := netip.ParseAddr(strings.TrimSpace(parts[i]))
			if err != nil {
				continue
			}
			addr = addr.Unmap()
			if !api.isTrustedProxy(addr) {
				return addr.String()
			}
		}
	}

	if ip := r.Header.Get("CF-Connecting-Ip"); net.ParseIP(ip) != nil {
		return ip
	}
	if ip := r.Header.Get("X-Real-Ip"); net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}

// remoteAddrIP returns the connection peer's IP ("" if unparseable).
func remoteAddrIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// No port (e.g. hand-built test requests).
		ip = r.RemoteAddr
	}
	if net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

// readPump drains incoming WebSocket frames and handles connection teardown.
//
// Must run in its own goroutine - concentrating all reads here ensures only
// one reader operates on the connection at a time.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	//#nosec G104 -- Disregard.
	//nolint:errcheck // ^
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			//#nosec G104 -- Disregard.
			//nolint:errcheck // ^
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
				websocket.CloseAbnormalClosure,
			) {
				logx.Error(
					err,
					logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					true,
				)
			}
			break
		}
	}
}

// writeServerMessage decodes and writes an outbound WebSocket message, rejecting unknown types.
func (c *Client) writeServerMessage(message []byte) {
	var msg apitype.WebSocketMessage
	if err := decode.Unmarshal("json", message, &msg); err != nil {
		logx.Error(
			"failed to unmarshal Message: "+err.Error(),
			logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
			true,
		)
		return
	}

	if msg.Page == "" || msg.Type == "" {
		return
	}

	switch msg.Type {
	case "VERSION", "WEBHOOK", "COMMAND", "SERVICE", "EDIT", "DELETE":
		if err := c.conn.WriteJSON(msg); err != nil {
			logx.Error(
				fmt.Sprintf(
					"Writing JSON to the WebSocket failed for %s: %s",
					msg.Type, err,
				),
				logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true,
			)
		}
	default:
		logx.Error(
			fmt.Sprintf(
				"Unknown Type (%q) in %q",
				msg.Type, string(message),
			),
			logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
			true,
		)
	}
}

// drainSendMessages writes all messages currently buffered in send without blocking.
func (c *Client) drainSendMessages() {
	for {
		select {
		default:
			return
		case queued, ok := <-c.send:
			if !ok {
				return
			}
			c.writeServerMessage(queued)
		}
	}
}

// writePump sends messages from the hub to the WebSocket connection.
//
// Must run in its own goroutine - concentrating all writes here ensures only
// one writer operates on the connection at a time.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			//#nosec G104 -- Disregard.
			//nolint:errcheck // ^
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel: log, close and return.
				logx.Verbose(
					"Closing the connection (writePump)",
					logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					true,
				)
				//nolint:errcheck // Best-effort close frame; the channel is gone.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.writeServerMessage(message)
			c.drainSendMessages()

		case <-ticker.C:
			//#nosec G104 -- Disregard.
			//nolint:errcheck // ^
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
