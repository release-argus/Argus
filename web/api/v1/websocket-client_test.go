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

//go:build unit

package v1

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestGetIP(t *testing.T) {
	// GIVEN: a request.
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		want       string
	}{
		{
			name: "CF-Connecting-Ip",
			want: "1.1.1.1",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-REAL-IP":        "2.2.2.2",
				"X-FORWARDED-FOR":  "3.3.3.3",
			},
			remoteAddr: "4.4.4.4:123",
		},
		{
			name: "X-Real-Ip",
			want: "2.2.2.2",
			headers: map[string]string{
				"X-REAL-IP":       "2.2.2.2",
				"X-FORWARDED-FOR": "3.3.3.3",
			},
			remoteAddr: "4.4.4.4:123",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3",
			},
			remoteAddr: "4.4.4.4:123",
			want:       "3.3.3.3",
		},
		{
			name:       "RemoteAddr",
			want:       "4.4.4.4",
			remoteAddr: "4.4.4.4:123",
		},
		{
			name:       "invalid RemoteAddr/SplitHostPort fail",
			want:       "",
			remoteAddr: "1111",
		},
		{
			name:       "invalid RemoteAddr/ParseIP fail",
			want:       "",
			remoteAddr: "1111:123",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			for header, val := range tc.headers {
				req.Header.Set(header, val)
			}
			req.RemoteAddr = tc.remoteAddr

			// WHEN: getIP is called on this request.
			got := getIP(req)

			prefix := fmt.Sprintf(
				"%s\ngetIP(%+v)",
				packageName, req,
			)

			// THEN: the function returns the correct result.
			if got != tc.want {
				t.Errorf(
					"%s value mismatch\ngot:  %v\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

type wsTestClient struct {
	client *Client
	conn   *websocket.Conn
	peer   *websocket.Conn
	server *httptest.Server
}

func setupWSTestClient(t *testing.T) *wsTestClient {
	t.Helper()

	// Create an upgrader for the test server.
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// Create test server.
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Fatalf(
						"%s\ncould not upgrade connection: %v",
						packageName, err,
					)
					return
				}

				// Echo messages back.
				for {
					messageType, message, err := ws.ReadMessage()
					if err != nil {
						return
					}
					if err := ws.WriteMessage(messageType, message); err != nil {
						return
					}
				}
			}),
	)

	// Create a WebSocket connection.
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf(
			"%s\ncould not open WebSocket connection: %v",
			packageName, err,
		)
	}

	// Create client.
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
		peer:   ws,
		server: server,
	}
}

// cleanup closes the [conn], [peer], and [server] channels of the [wsTestClient]
func (w *wsTestClient) cleanup(t *testing.T) {
	t.Helper()

	_ = w.conn.Close()
	_ = w.peer.Close()
	w.server.Close()
}

// closePeer will close the [peer] channel of this [wsTestClient] with the given code.
func (w *wsTestClient) closePeer(t *testing.T, code int) {
	t.Helper()

	if code == 0 {
		_ = w.peer.Close()
		return
	}

	_ = w.peer.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, ""),
	)
}

func TestClient_ReadPump__pongHandler(t *testing.T) {
	// GIVEN: a Hub with a registered client and a running readPump.
	wsTest := setupWSHubClient(t)
	t.Cleanup(func() { wsTest.cleanup(t) })
	go wsTest.client.readPump()

	prefix := fmt.Sprintf("%s\nClient.readPump()", packageName)

	// WHEN: the client sends a Ping
	if err := wsTest.client.conn.WriteControl(
		websocket.PingMessage,
		[]byte{},
		time.Now().Add(time.Second),
	); err != nil {
		t.Fatalf("%s failed to write ping: %v", prefix, err)
	}
	time.Sleep(100 * time.Millisecond)

	// THEN: the connection remains alive — pong handler ran without error.
	if !wsTest.client.hub.clients[wsTest.client] {
		t.Errorf("%s client should still be registered to hub after pong", prefix)
	}
}

func TestClient_ReadPump__disconnects(t *testing.T) {
	tests := []struct {
		name            string
		peerMessages    []string
		peerCloseCode   int
		closeClientConn bool
		stdoutRegex     string
	}{
		{
			name: "message exceeds read limit", // [websocket.ErrReadLimit]
			peerMessages: []string{
				strings.Repeat("a", maxMessageSize+1),
			},
			stdoutRegex: `^$`,
		},
		{
			name:          "expected peer close",
			peerCloseCode: websocket.CloseGoingAway,
			stdoutRegex:   `^$`,
		},
		{
			name:          "unexpected peer close",
			peerCloseCode: websocket.CloseNormalClosure,
			stdoutRegex:   `^ERROR: .*WebSocket.*127.0.0.1.*`,
		},
		{
			name:            "abrupt client close unregisters from hub",
			closeClientConn: true,
			stdoutRegex:     `$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// GIVEN: a Hub with a registered client.
			wsTest := setupWSHubClient(t)
			t.Cleanup(func() { wsTest.cleanup(t) })

			prefix := fmt.Sprintf("%s\nClient.readPump()", packageName)

			// AND: a readPump is running for this client.
			go wsTest.client.readPump()

			// WHEN: the connection is disrupted.
			for _, msg := range tc.peerMessages {
				if err := wsTest.peer.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					t.Fatalf("%s failed to write message from peer: %v", prefix, err)
				}
			}
			switch {
			case tc.peerCloseCode != 0:
				// AND: the peer closes with a given code.
				wsTest.closePeer(t, tc.peerCloseCode)
			case tc.closeClientConn:
				// AND: the connection is closed without a close message.
				_ = wsTest.conn.Close()
			}

			// THEN: the client is unregistered and removed from the hub.
			waitForClientUnregistered(t, wsTest.client.hub, wsTest.client)
			// AND: stdout matches.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestClient_WriteServerMessage(t *testing.T) {
	tests := []struct {
		name         string
		messages     []string
		wantMessages []string
		closeConn    bool
		stdoutRegex  string
	}{
		{
			name: "skips messages without type or page",
			messages: []string{
				`{"version":null,"type":"VERSION"}`,
				`{"version":null}`,
				`{}`,
			},
			wantMessages: []string{},
			stdoutRegex:  `^$`,
		},
		{
			name: "writes server messages",
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
				`{"version":null,"type":"COMMAND","page":"home"}`,
				`{"version":null,"type":"SERVICE","page":"home"}`,
				`{"version":null,"type":"EDIT","page":"home"}`,
				`{"version":null,"type":"DELETE","page":"home"}`,
			},
			wantMessages: []string{
				`{"page":"home","type":"VERSION"}` + "\n",
				`{"page":"home","type":"WEBHOOK"}` + "\n",
				`{"page":"home","type":"COMMAND"}` + "\n",
				`{"page":"home","type":"SERVICE"}` + "\n",
				`{"page":"home","type":"EDIT"}` + "\n",
				`{"page":"home","type":"DELETE"}` + "\n",
			},
			stdoutRegex: `^$`,
		},
		{
			name: "rejects invalid message type",
			messages: []string{
				`{"version":null,"type":"INVALID","page":"home"}`,
			},
			wantMessages: []string{},
			stdoutRegex: test.TrimYAML(`
				^ERROR:.*Unknown Type.*"INVALID".*`,
			),
		},
		{
			name: "rejects invalid JSON",
			messages: []string{
				`{"invalid`,
			},
			wantMessages: []string{},
			stdoutRegex:  `^ERROR: .*, failed to unmarshal Message: .*`,
		},
		{
			name: "logs write failure for server message",
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			wantMessages: []string{},
			closeConn:    true,
			stdoutRegex:  `^ERROR: .*, Writing JSON to the WebSocket failed for VERSION`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			prefix := fmt.Sprintf("%s\nClient.writeServerMessage()", packageName)

			// GIVEN: a client with a WebSocket connection.
			wsTest := setupWSTestClient(t)
			t.Cleanup(func() { wsTest.cleanup(t) })

			if tc.closeConn {
				// AND: the client connection is closed.
				if err := wsTest.conn.Close(); err != nil {
					t.Fatalf(
						"%s failed to close connection: %v",
						prefix, err,
					)
				}
			}

			// WHEN: writeServerMessage is called for each message.
			for _, msg := range tc.messages {
				wsTest.client.writeServerMessage([]byte(msg))
			}

			// THEN: the expected messages are written to the WebSocket connection.
			got := readConnMessages(t, wsTest.conn, prefix, len(tc.wantMessages))
			assertConnMessages(t, prefix, got, tc.wantMessages)

			// AND: stdout matches.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestClient_DrainSendMessages(t *testing.T) {
	tests := []struct {
		name         string
		queued       []string
		writeFirst   string
		closeSend    bool
		wantMessages []string
		stdoutRegex  string
	}{
		{
			name:        "returns immediately when send is empty",
			stdoutRegex: `^$`,
		},
		{
			name: "drains buffered messages after first message written",
			queued: []string{
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
			},
			writeFirst: `{"version":null,"type":"VERSION","page":"home"}`,
			wantMessages: []string{
				`{"page":"home","type":"VERSION"}` + "\n",
				`{"page":"home","type":"WEBHOOK"}` + "\n",
			},
			stdoutRegex: `^$`,
		},
		{
			name: "queued messages sent",
			queued: []string{
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			writeFirst: "",
			wantMessages: []string{
				`{"page":"home","type":"WEBHOOK"}` + "\n",
				`{"page":"home","type":"VERSION"}` + "\n",
			},
			stdoutRegex: `^$`,
		},
		{
			name: "drains buffered messages on closed send channel",
			queued: []string{
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
			},
			writeFirst: `{"version":null,"type":"VERSION","page":"home"}`,
			closeSend:  true,
			wantMessages: []string{
				`{"page":"home","type":"VERSION"}` + "\n",
				`{"page":"home","type":"WEBHOOK"}` + "\n",
			},
			stdoutRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// GIVEN: a client with a WebSocket connection.
			wsTest := setupWSTestClient(t)
			t.Cleanup(func() { wsTest.cleanup(t) })

			prefix := fmt.Sprintf("%s\nClient.drainSendMessages()", packageName)

			// AND: messages are queued to send to the client.
			for _, msg := range tc.queued {
				wsTest.client.send <- []byte(msg)
			}

			// AND: the client channel is closed.
			if tc.closeSend {
				close(wsTest.client.send)
			}

			// AND: a message is written to the client.
			if tc.writeFirst != "" {
				wsTest.client.writeServerMessage([]byte(tc.writeFirst))
			}

			// WHEN: drainSendMessages is called.
			wsTest.client.drainSendMessages()

			// THEN: expected messages are written to the WebSocket connection.
			time.Sleep(250 * time.Millisecond)
			got := readConnMessages(t, wsTest.conn, prefix, len(tc.wantMessages))
			assertConnMessages(t, prefix, got, tc.wantMessages)

			// AND: the queued messages channel is emptied.
			if got := len(wsTest.client.send); got != 0 {
				t.Errorf(
					"%s send channel length mismatch\ngot:  %d\nwant: 0",
					prefix, got,
				)
			}

			// AND: stdout matches.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestClient_WritePump__messages(t *testing.T) {
	tests := []struct {
		name         string
		messages     []string
		wantMessages []string
		stdoutRegex  string
	}{
		{
			name: "sends and drains queued server messages",
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
				`{"version":null,"type":"WEBHOOK","page":"home"}`,
			},
			wantMessages: []string{
				`{"page":"home","type":"VERSION"}` + "\n",
				`{"page":"home","type":"WEBHOOK"}` + "\n",
			},
			stdoutRegex: `^$`,
		},
		{
			name: "rejects invalid message type",
			messages: []string{
				`{"version":null,"type":"INVALID","page":"home"}`,
			},
			wantMessages: []string{},
			stdoutRegex: test.TrimYAML(`
				^ERROR: .*Unknown Type.*INVALID.*`,
			),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			// GIVEN: a Hub with a registered client.
			wsTest := setupWSHubClient(t)
			t.Cleanup(func() { wsTest.cleanup(t) })

			prefix := fmt.Sprintf("%s\nClient.writePump()", packageName)

			// AND: a writePump is running for this client.
			go wsTest.client.writePump()

			// WHEN: messages are sent through the client's send channel.
			for _, msg := range tc.messages {
				wsTest.client.send <- []byte(msg)
			}

			// THEN: messages are written or rejected as expected.
			time.Sleep(250 * time.Millisecond)
			got := readConnMessages(t, wsTest.conn, prefix, len(tc.wantMessages))
			assertConnMessages(t, prefix, got, tc.wantMessages)

			// AND: stdout matches.
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func TestClient_WritePump__connection(t *testing.T) {
	tests := []struct {
		name                string
		messages            []string
		closeSend           bool
		closeConnBeforePump bool
		closeConnAfterPump  bool
		shortenPing         bool
		stdoutRegex         string
	}{
		{
			name:        "closes on empty send channel",
			closeSend:   true,
			stdoutRegex: `VERBOSE: .*Closing the connection \(writePump\)`,
		},
		{
			name: "logs write failure when connection closed",
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			closeConnBeforePump: true,
			stdoutRegex:         `^ERROR:.*Writing JSON to the WebSocket failed for VERSION`,
		},
		{
			name:        "sends ping frames",
			shortenPing: true,
			stdoutRegex: `^$`,
		},
		{
			name:               "ping write failure exits writePump",
			shortenPing:        true,
			closeConnAfterPump: true,
			stdoutRegex:        `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're using stdout.
			releaseStdout := test.CaptureLog(t, logx.Default())

			if tc.shortenPing {
				oldPingPeriod := pingPeriod
				pingPeriod = 10 * time.Millisecond
				t.Cleanup(func() { pingPeriod = oldPingPeriod })
			}

			// GIVEN: a Hub with a registered client.
			wsTest := setupWSHubClient(t)
			t.Cleanup(func() { wsTest.cleanup(t) })

			prefix := fmt.Sprintf("%s\nClient.writePump()", packageName)

			// AND: the client connection is closed before the writePump starts.
			if tc.closeConnBeforePump {
				if err := wsTest.conn.Close(); err != nil {
					t.Fatalf("%s failed to close connection: %v", prefix, err)
				}
			}

			// AND: the messages are sent by the client.
			for _, msg := range tc.messages {
				wsTest.client.send <- []byte(msg)
			}

			if tc.closeConnBeforePump {
				// WHEN: the send channel is closed before the writePump starts.
				if tc.closeSend {
					close(wsTest.client.send)
				}

				// AND: the writePump starts.
				go wsTest.client.writePump()
			} else {
				// WHEN: the writePump starts.
				go wsTest.client.writePump()

				// AND: the send channel is closed.
				if tc.closeSend {
					close(wsTest.client.send)
				}

				// AND: the connection is closed after the writePump starts.
				if tc.closeConnAfterPump {
					if err := wsTest.conn.Close(); err != nil {
						t.Fatalf("%s failed to close connection: %v", prefix, err)
					}
				}
			}

			// AND: stdout matches.
			time.Sleep(300 * time.Millisecond)
			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf(
					"%s stdout mismatch\ngot:  %q\nwant: %q",
					prefix, stdout, tc.stdoutRegex,
				)
			}
		})
	}
}

func readConnMessages(t *testing.T, conn *websocket.Conn, prefix string, count int) []string {
	t.Helper()

	got, err := tryReadConnMessages(conn, count)
	if err != nil {
		t.Fatalf("%s %v", prefix, err)
	}
	return got
}

func tryReadConnMessages(conn *websocket.Conn, count int) ([]string, error) {
	received := make([]string, 0, count)
	for i := range count {
		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			return nil, fmt.Errorf("failed to set read deadline: %w", err)
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read message [%d] from WebSocket connection: %w",
				i, err,
			)
		}
		received = append(received, string(message))
	}
	return received, nil
}

func assertConnMessages(t *testing.T, prefix string, got, want []string) {
	t.Helper()

	if gotLen, wantLen := len(got), len(want); gotLen != wantLen {
		t.Errorf("%s message count mismatch\ngot:  %d\nwant: %d", prefix, gotLen, wantLen)
	}
	for i, wantMsg := range want {
		if i >= len(got) {
			break
		}
		if gotMsg := got[i]; gotMsg != wantMsg {
			t.Errorf(
				"%s mismatch on message [%d]\ngot:  %q\nwant: %q",
				prefix, i, gotMsg, wantMsg,
			)
		}
	}
}

// setupWSHubClient returns a registered client with a running hub.
func setupWSHubClient(t *testing.T) *wsTestClient {
	t.Helper()

	// GIVEN: a test client/hub.
	wsTest := setupWSTestClient(t)
	t.Cleanup(func() { wsTest.cleanup(t) })
	go wsTest.client.hub.Run()
	wsTest.client.hub.register <- wsTest.client
	time.Sleep(100 * time.Millisecond)

	return wsTest
}

// waitForClientUnregistered fails if the client is not unregistered from the hub quicker than timeout.
func waitForClientUnregistered(t *testing.T, hub *Hub, client *Client) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !hub.clients[client] {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s\ntimed out waiting for client to unregister from hub", packageName)
}

func TestServeWs(t *testing.T) {
	tests := []struct {
		name           string
		dialHeaders    map[string]string
		wantStatus     int
		wantRegistered bool
		wantIP         string
		wantSendCap    int
	}{
		{
			name:           "upgrades and registers client",
			wantStatus:     http.StatusSwitchingProtocols,
			wantRegistered: true,
			wantIP:         "127.0.0.1",
			wantSendCap:    256,
		},
		{
			name:           "uses CF-Connecting-Ip",
			dialHeaders:    map[string]string{"CF-Connecting-IP": "2.2.2.2"},
			wantStatus:     http.StatusSwitchingProtocols,
			wantRegistered: true,
			wantIP:         "2.2.2.2",
			wantSendCap:    256,
		},
		{
			name:           "uses X-Real-Ip",
			dialHeaders:    map[string]string{"X-Real-Ip": "3.3.3.3"},
			wantStatus:     http.StatusSwitchingProtocols,
			wantRegistered: true,
			wantIP:         "3.3.3.3",
			wantSendCap:    256,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// GIVEN: a Hub.
			hub := NewHub()
			go hub.Run()

			// AND: a HTTP server with a WebSocket endpoint.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ServeWs(hub, w, r)
			}))
			t.Cleanup(server.Close)

			prefix := fmt.Sprintf("%s\nServeWs()", packageName)

			// WHEN: a WebSocket client connects.
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			dialHeader := http.Header{}
			for key, value := range tc.dialHeaders {
				dialHeader.Set(key, value)
			}
			clientConn, resp, err := websocket.DefaultDialer.Dial(wsURL, dialHeader)
			if err != nil {
				t.Fatalf("%s failed to dial WebSocket: %v", prefix, err)
			}
			t.Cleanup(func() { _ = clientConn.Close() })

			// THEN: the upgrade succeeds.
			if got, want := resp.StatusCode, tc.wantStatus; got != want {
				t.Errorf("%s status mismatch\ngot:  %d\nwant: %d", prefix, got, want)
			}
			time.Sleep(time.Second)

			// AND: the client is registered on the Hub.
			registered := hubClientForTest(t, hub)
			if tc.wantRegistered && registered == nil {
				t.Fatalf("%s expected a registered client", prefix)
			}
			if !tc.wantRegistered && registered != nil {
				t.Fatalf("%s expected no registered client, got %p", prefix, registered)
			}
			if registered == nil {
				return
			}

			// AND: the client's data is filled correctly.
			fieldTests := []test.FieldAssertion{
				{Name: "IP", Got: registered.ip, Want: tc.wantIP, Mode: test.CompareEqual},
				{Name: "Send capacity", Got: cap(registered.send), Want: tc.wantSendCap, Mode: test.CompareEqual},
				{Name: "Hub", Got: registered.hub, Want: hub, Mode: test.CompareSamePointer},
				{Name: "Conn", Got: registered.conn, Want: nil, Mode: test.CompareDifferentPointer},
			}
			if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestServeWs__plain_HTTP(t *testing.T) {
	// GIVEN: a Hub.
	hub := NewHub()
	go hub.Run()

	// AND: a HTTP server with a WebSocket endpoint.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	}))
	t.Cleanup(server.Close)

	prefix := fmt.Sprintf("%s\nServeWs()", packageName)

	// AND: a plain HTTP request to the WebSocket endpoint.
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	// WHEN: a plain HTTP request is served.
	ServeWs(hub, rec, req)

	// THEN: the upgrade fails.
	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Errorf("%s status mismatch\ngot:  %d\nwant: %d", prefix, got, want)
	}

	time.Sleep(100 * time.Millisecond)

	// AND: no client is registered.
	if registered := hubClientForTest(t, hub); registered != nil {
		t.Errorf("%s expected no registered client, got %p", prefix, registered)
	}
}

func hubClientForTest(t *testing.T, hub *Hub) *Client {
	t.Helper()

	if len(hub.clients) == 0 {
		return nil
	}

	if len(hub.clients) > 1 {
		t.Fatalf("%s\nexpected one registered client, got %d", packageName, len(hub.clients))
	}

	for client := range hub.clients {
		return client
	}
	return nil
}
