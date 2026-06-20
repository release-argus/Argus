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
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// ServeWs upgrades an HTTP connection to WebSocket and registers the client with the hub.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	conn.RemoteAddr()
	client := &Client{
		hub:  hub,
		ip:   getIP(r),
		conn: conn,
		send: make(chan []byte, 256),
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
}

// getIP returns the client IP from proxy headers or RemoteAddr.
func getIP(r *http.Request) (ip string) {
	// Get IP from the CF-Connecting-Ip header.
	ip = r.Header.Get("CF-Connecting-Ip")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return
	}

	// Get IP from the X-Real-Ip header.
	ip = r.Header.Get("X-Real-Ip")
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return
	}

	// Get IP from X-Forwarded-For header.
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")
	for _, ip = range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return
		}
	}

	// Get IP from RemoteAddr.
	var err error
	ip, _, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return
	}

	return ""
}

// serverMessageCheck checks whether the message came from the server.
type serverMessageCheck struct {
	Version int `json:"version"`
}

// readPump reads incoming WebSocket frames and handles connection teardown.
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
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseGoingAway,
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

		if logx.IsLevel("DEBUG") {
			logx.Debug(
				fmt.Sprintf("READ %q", message),
				logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true,
			)
		}

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		// Ensure it meets client message format.
		var validation serverMessageCheck
		if err := decode.Unmarshal("json", message, &validation); err != nil || validation.Version != 1 {
			logx.Warn(
				fmt.Sprintf("Invalid message (missing/invalid version key) - %q", message),
				logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true,
			)
			continue
		}

		c.send <- message
	}
}

// writeServerMessage decodes and writes an outbound WebSocket message, rejecting unknown types.
func (c *Client) writeWebSocketMessage(message []byte) {
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

	// If message came from the server (doesn't use version).
	if msg.Version == nil {
		switch msg.Type {
		case "VERSION", "WEBHOOK", "COMMAND", "SERVICE", "EDIT", "DELETE":
			if err := c.conn.WriteJSON(msg); err != nil {
				logx.Error(
					fmt.Sprintf(
						"Writing JSON to the websocket failed for %s\n%s",
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
		return
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		logx.Error(
			fmt.Sprintf("WriteJSON for the queued chat messages\n%s", err),
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
			c.writeWebSocketMessage(queued)
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
				// The hub closed the channel.
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					logx.Verbose(
						fmt.Sprintf("Closing the connection (writePump) - %q", err),
						logx.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true,
					)
					return
				}
			}

			c.writeWebSocketMessage(message)
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
