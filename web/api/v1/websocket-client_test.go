// Copyright [2024] [Argus]
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

//go:build unit

package v1

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
)

func TestGetIP(t *testing.T) {
	// GIVEN a request
	tests := map[string]struct {
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		"CF-Connecting-Ip": {
			want: "1.1.1.1",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-REAL-IP":        "2.2.2.2",
				"X-FORWARDED-FOR":  "3.3.3.3"},
			remoteAddr: "4.4.4.4:123"},
		"X-Real-Ip": {
			want: "2.2.2.2",
			headers: map[string]string{
				"X-REAL-IP":       "2.2.2.2",
				"X-FORWARDED-FOR": "3.3.3.3"},
			remoteAddr: "4.4.4.4:123"},
		"X-Forwarded-For": {
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3"},
			remoteAddr: "4.4.4.4:123",
			want:       "3.3.3.3"},
		"RemoteAddr": {
			want:       "4.4.4.4",
			remoteAddr: "4.4.4.4:123"},
		"Invalid RemoteAddr (SplitHostPort fail)": {
			want:       "",
			remoteAddr: "1111"},
		"Invalid RemoteAddr (ParseIP fail)": {
			want:       "",
			remoteAddr: "1111:123"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			for header, val := range tc.headers {
				req.Header.Set(header, val)
			}
			req.RemoteAddr = tc.remoteAddr

			// WHEN getIP is called on this request
			got := getIP(req)

			// THEN the function returns the correct result
			if got != tc.want {
				t.Errorf("want: %q\ngot:  %v",
					tc.want, got)
			}
		})
	}
}

type wsTestClient struct {
	client *Client
	conn   *websocket.Conn
	server *httptest.Server
}

func setupWSTestClient(t *testing.T) *wsTestClient {
	t.Helper()

	// Create an upgrader for the test server
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("could not upgrade connection: %v", err)
			return
		}
		defer ws.Close()

		// Echo messages back
		for {
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				break
			}
			if err := ws.WriteMessage(messageType, message); err != nil {
				break
			}
		}
	}))

	// Create WebSocket connection
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not open websocket connection: %v", err)
	}

	// Create client
	hub := NewHub()
	client := &Client{
		hub:  hub,
		ip:   "127.0.0.1",
		conn: ws,
		send: make(chan []byte, 256),
	}

	return &wsTestClient{
		client: client,
		conn:   ws,
		server: server,
	}
}

func (w *wsTestClient) cleanup() {
	w.conn.Close()
	w.server.Close()
}

func TestClient_writePump(t *testing.T) {
	tests := map[string]struct {
		messages     []string
		wantMessages []string
		closeClient  bool
		stdoutRegex  string
	}{
		"No type/page": {
			messages: []string{
				`{"version":null,"type":"VERSION"}`,
				`{"version":null}`,
				`{}`,
			},
			wantMessages: []string{},
			stdoutRegex:  `^$`,
		},
		"valid message": {
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			wantMessages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			stdoutRegex: `^$`,
		},
		"valid messages": {
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
				`{"version":null,"type":"COMMAND","page":"home"}`,
				`{"version":null,"type":"SERVICE","page":"home"}`,
				`{"version":null,"type":"EDIT","page":"home"}`,
				`{"version":null,"type":"DELETE","page":"home"}`,
			},
			wantMessages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
				`{"version":null,"type":"COMMAND","page":"home"}`,
				`{"version":null,"type":"SERVICE","page":"home"}`,
				`{"version":null,"type":"EDIT","page":"home"}`,
				`{"version":null,"type":"DELETE","page":"home"}`,
			},
			stdoutRegex: `^$`,
		},
		"invalid message type": {
			messages: []string{
				`{"version":null,"type":"INVALID","page":"home"}`,
			},
			stdoutRegex: test.TrimYAML(`
				^ERROR:.*Unknown TYPE.*
				.*INVALID.*`),
		},
		"invalid json": {
			messages: []string{
				`{"invalid`,
			},
			wantMessages: []string{},
			stdoutRegex: test.TrimYAML(`
				^ERROR:.*Message failed to unmarshal.*`),
		},
		"close client": {
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			wantMessages: []string{},
			closeClient:  true,
			stdoutRegex: test.TrimYAML(`
				^ERROR:.*Message failed to unmarshal.*`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			// Setup test client
			wsTest := setupWSTestClient(t)
			defer wsTest.cleanup()

			// Start writePump
			go wsTest.client.writePump()

			// Create channel to track received messages
			receivedMessages := make([]string, 0)
			done := make(chan bool)

			// Start goroutine to read messages
			go func() {
				for _, _ = range tc.messages {
					_, message, err := wsTest.conn.ReadMessage()
					if err != nil {
						if !tc.closeClient {
							t.Logf("unexpected error reading message: %v", err)
						}
						break
					}
					receivedMessages = append(receivedMessages, string(message))
					if len(receivedMessages) == len(tc.messages) {
						done <- true
						return
					}
				}
				done <- true
			}()

			// Send messages through the client's send channel
			for _, msg := range tc.messages {
				wsTest.client.send <- []byte(msg)
			}

			if tc.closeClient {
				close(wsTest.client.send)
			}

			// Wait for messages or timeout
			select {
			case <-done:
				// Success case
			case <-time.After(2 * time.Second):
				if !tc.closeClient && len(receivedMessages) != len(tc.wantMessages) {
					t.Errorf("timeout waiting for messages. Got %d messages, want %d",
						len(receivedMessages), len(tc.messages))
				}
			}

			// Verify stdout
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout did not match regex %q:\n%s", tc.stdoutRegex, stdout)
			}
		})
	}
}

func TestClient_readPump(t *testing.T) {
	tests := map[string]struct {
		messages        []string
		closeConnection bool
		stdoutRegex     string
		wantMessages    []string
	}{
		"valid message": {
			messages: []string{
				`{"version":1,"type":"test","page":"home"}`,
			},
			stdoutRegex: `^$`,
			wantMessages: []string{
				`{"version":1,"type":"test","page":"home"}`,
			},
		},
		"invalid json message": {
			messages: []string{`{"invalid`},
			stdoutRegex: test.TrimYAML(`
				^WARNING:.*Invalid message.*
				\{"invalid`),
		},
		"message missing version": {
			messages: []string{`{"type":"test","page":"home"}`},
			stdoutRegex: test.TrimYAML(`
				^WARNING:.*Invalid message.*
				\{.*\}`),
		},
		"connection closed": {
			messages:        []string{`{"version":1,"type":"test","page":"home"}`},
			closeConnection: true,
			wantMessages:    []string{},
			stdoutRegex:     `^$`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout
			releaseStdout := test.CaptureStdout()

			// Setup test client
			wsTest := setupWSTestClient(t)
			defer wsTest.cleanup()

			// Create channels to coordinate test
			done := make(chan bool)
			receivedMessages := make([]string, 0)

			// Start goroutine to collect messages from send channel
			go func() {
				for msg := range wsTest.client.send {
					receivedMessages = append(receivedMessages, string(msg))
					if len(receivedMessages) == len(tc.wantMessages) {
						done <- true
						return
					}
				}
			}()

			// Start the readPump
			go wsTest.client.readPump()

			// Send test messages
			for _, msg := range tc.messages {
				err := wsTest.client.conn.WriteMessage(websocket.TextMessage, []byte(msg))
				if err != nil {
					t.Fatalf("failed to write message: %v", err)
				}
			}

			if tc.closeConnection {
				wsTest.conn.Close()
			}

			// Wait for messages or timeout
			select {
			case <-done:
				// Success case
			case <-time.After(2 * time.Second):
				if len(receivedMessages) != len(tc.wantMessages) {
					t.Errorf("timeout waiting for messages. Got %d messages, want %d",
						len(receivedMessages), len(tc.wantMessages))
				}
			}

			// Verify received messages
			if len(receivedMessages) != len(tc.wantMessages) {
				t.Errorf("got %d messages, want %d", len(receivedMessages), len(tc.wantMessages))
			}
			for i, want := range tc.wantMessages {
				if i >= len(receivedMessages) {
					break
				}
				if receivedMessages[i] != want {
					t.Errorf("message %d:\ngot:  %s\nwant: %s", i, receivedMessages[i], want)
				}
			}

			// THEN verify the results
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout did not match regex %q:\n%s", tc.stdoutRegex, stdout)
			}
		})
	}
}
