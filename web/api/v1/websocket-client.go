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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/util"
	api_type "github.com/release-argus/Argus/web/api/types"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
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

	// Allow all connections
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// The API.
	api *API

	// The WebSocket hub.
	hub *Hub

	// The client's IP
	ip string

	// The WebSocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Lock to prevent concurrent write panics
	lock sync.Mutex
}

func getIP(r *http.Request) (ip string) {
	// Get IP from the CF-Connecting-Ip header
	ip = r.Header.Get("CF-Connecting-Ip")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return
	}

	// Get IP from the X-Real-Ip header
	ip = r.Header.Get("X-Real-Ip")
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return
	}

	// Get IP from X-Forwarded-For header
	ips := r.Header.Get("X-Forwarded-For")
	splitIps := strings.Split(ips, ",")
	for _, ip = range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return
		}
	}

	// Get IP from RemoteAddr
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

type serverMessage struct {
	Version int `json:"version"`
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		c.api.Log.Verbose(
			fmt.Sprintf("Closing the websocket connection failed (readPump)\n%s", util.ErrorToString(err)),
			util.LogFrom{},
			err != nil,
		)
	}()
	c.conn.SetReadLimit(maxMessageSize)
	//#nosec G104 -- Disregard
	//nolint:errcheck // ^
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			//#nosec G104 -- Disregard
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
				log.Printf("error: %v\n", err)
			}
			break
		}

		c.api.Log.Debug(
			fmt.Sprintf("READ %s", message),
			util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
			true,
		)

		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		// Check it's not trying to be a server message by omitting the version key
		var validation serverMessage
		err = json.Unmarshal(message, &validation)
		if err != nil {
			c.api.Log.Warn(
				fmt.Sprintf("Invalid message (missing/invalid version key)\n%s", message),
				util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
				true,
			)
			continue
		}

		c.send <- message
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		c.api.Log.Verbose(
			fmt.Sprintf("Closing the connection\n%s", util.ErrorToString(err)),
			util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
			true,
		)
	}()
	for {
		select {
		case message, ok := <-c.send:
			//#nosec G104 -- Disregard
			//nolint:errcheck // ^
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// The hub closed the channel.
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.api.Log.Verbose(
					fmt.Sprintf("Closing the connection (writePump)\n%s", util.ErrorToString(err)),
					util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					true,
				)
				return
			}

			var msg api_type.WebSocketMessage
			err := json.Unmarshal(message, &msg)
			if err != nil {
				c.api.Log.Error(
					fmt.Sprintf("Message failed to unmarshal %s", util.ErrorToString(err)),
					util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					true,
				)
				continue
			}

			if msg.Page == "" || msg.Type == "" {
				continue
			}
			// If message is from the server (doesn't use version)
			if msg.Version == nil {
				switch msg.Type {
				case "VERSION", "WEBHOOK", "COMMAND", "SERVICE", "EDIT", "DELETE":
					err := c.conn.WriteJSON(msg)
					c.api.Log.Error(
						fmt.Sprintf("Writing JSON to the websocket failed for %s\n%s", msg.Type, util.ErrorToString(err)),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						err != nil,
					)
				default:
					c.api.Log.Error(
						fmt.Sprintf("Unknown TYPE %q\nFull message: %s", msg.Type, string(message)),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true,
					)
					continue
				}
			} else {
				// Message is from client (`msg.Version` specified)
				switch msg.Page {
				case "APPROVALS":
					switch msg.Type {
					case "VERSION":
						// Approval/Skip
						c.api.wsServiceAction(c, msg)
					case "ACTIONS":
						// Get Command data for a service
						c.api.wsCommand(c, msg)
						// Get WebHook data for a service
						c.api.wsWebHook(c, msg)
					case "INIT":
						// Get all Service data
						c.api.wsServiceInit(c)
					default:
						c.api.Log.Error(
							fmt.Sprintf("Unknown APPROVALS Type %q\nFull message: %s", msg.Type, string(message)),
							util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
							true,
						)
						continue
					}
				case "RUNTIME_BUILD":
					switch msg.Type {
					case "INIT":
						c.api.wsStatus(c)
					default:
						c.api.Log.Error(
							fmt.Sprintf("Unknown RUNTIME_BUILD Type %q\nFull message: %s", msg.Type, string(message)),
							util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
							true,
						)
					}
				case "FLAGS":
					switch msg.Type {
					case "INIT":
						c.api.wsFlags(c)
					default:
						c.api.Log.Error(
							fmt.Sprintf("Unknown FLAGS Type %q\nFull message: %s", msg.Type, string(message)),
							util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
							true,
						)
						continue
					}
				case "CONFIG":
					switch msg.Type {
					case "INIT":
						c.api.wsConfigSettings(c)
						c.api.wsConfigDefaults(c)
						c.api.wsConfigNotify(c)
						c.api.wsConfigWebHook(c)
						c.api.wsConfigService(c)
					default:
						c.api.Log.Error(
							fmt.Sprintf("Unknown CONFIG Type %q\nFull message: %s", msg.Type, string(message)),
							util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
							true,
						)
						continue
					}
				default:
					c.api.Log.Error(
						fmt.Sprintf("Unknown PAGE %q\nFull message: %s", msg.Page, string(message)),
						util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
						true,
					)
					continue
				}
			}

			// Send all queued chat messages.
			n := len(c.send)
			for i := 0; i < n; i++ {
				err := c.conn.WriteJSON(<-c.send)
				c.api.Log.Error(
					fmt.Sprintf("WriteJSON for the queued chat messages\n%s\n", util.ErrorToString(err)),
					util.LogFrom{Primary: "WebSocket", Secondary: c.ip},
					err != nil,
				)
			}

		case <-ticker.C:
			//#nosec G104 -- Disregard
			//nolint:errcheck // ^
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(api *API, hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	conn.RemoteAddr()
	client := &Client{
		api:  api,
		hub:  hub,
		ip:   getIP(r),
		conn: conn,
		send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
