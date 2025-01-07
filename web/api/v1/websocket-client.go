// Copyright [2025] [Argus]
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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/util"
	apitype "github.com/release-argus/Argus/web/api/types"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to the peer at this interval. Must occur before pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all connections.
	CheckOrigin: func(_ *http.Request) bool { return true },
}

// Client connects the websocket and the hub.
type Client struct {
	// The WebSocket hub.
	hub *Hub

	// The client's IP.
	ip string

	// The WebSocket connection.
	conn *websocket.Conn

	// A buffered channel of outbound messages.
	send chan []byte
}

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

// readPump reads messages from the websocket connection to the hub.
//
// The application runs readPump in a separate goroutine for each connection.
// It ensures only one reader operates on a connection by running all
// reads in this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		if err := c.conn.Close(); err != nil {
			jLog.Verbose(
				fmt.Sprintf("Closing the websocket connection failed (readPump)\n%s",
					err),
				util.LogFrom{},
				true)
		}
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
				jLog.Error(err,
					util.LogFrom{Primary: "WebSocket", Secondary: c.ip}, true)
			}
			break
		}

		if jLog.IsLevel("DEBUG") {
			jLog.Debug(
				fmt.Sprintf("READ %s", message),
				util.LogFrom{Primary: "WebSocket", Secondary: c.ip}, true)
		}

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		// Ensure it meets client message format.
		var validation serverMessageCheck
		if err := json.Unmarshal(message, &validation); err != nil || validation.Version != 1 {
			jLog.Warn(
				fmt.Sprintf("Invalid message (missing/invalid version key)\n%s", message),
				util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true,
			)
			continue
		}

		c.send <- message
	}
}

// writePump sends messages from the hub to the websocket connection.
//
// The application starts a separate goroutine for each connection to run writePump.
// It ensures only one writer handles the connection by executing all writes in this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			jLog.Verbose(
				fmt.Sprintf("Closing the connection\n%s",
					err),
				util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true)
		}
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
					jLog.Verbose(
						fmt.Sprintf("Closing the connection (writePump)\n%s",
							err),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true)
					return
				}
			}

			var msg apitype.WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				jLog.Error(
					fmt.Sprintf("Message failed to unmarshal %s",
						err),
					util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					true)
				continue
			}

			if msg.Page == "" || msg.Type == "" {
				continue
			}
			// If message came from the server (doesn't use version).
			if msg.Version == nil {
				switch msg.Type {
				case "VERSION", "WEBHOOK", "COMMAND", "SERVICE", "EDIT", "DELETE":
					if err := c.conn.WriteJSON(msg); err != nil {
						jLog.Error(
							fmt.Sprintf("Writing JSON to the websocket failed for %s\n%s",
								msg.Type, err),
							util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
							true)
					}
				default:
					jLog.Error(
						fmt.Sprintf("Unknown TYPE %q\nFull message: %s", msg.Type, string(message)),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true)
					continue
				}
			}

			// Send all queued chat messages.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if err := c.conn.WriteJSON(<-c.send); err != nil {
					jLog.Error(
						fmt.Sprintf("WriteJSON for the queued chat messages\n%s\n",
							err),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true)
				}
			}

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

// ServeWs handles websocket requests from the peer.
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
		send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
