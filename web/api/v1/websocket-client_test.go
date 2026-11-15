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
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
)

func TestServeWs(t *testing.T) {
	tests := []struct {
		name        string
		dialHeaders map[string]string
		trustPeer   bool
		wantStatus  int
		wantIP      string
		wantSendCap int
	}{
		{
			name:        "upgrades and registers client",
			wantStatus:  http.StatusSwitchingProtocols,
			wantIP:      "127.0.0.1",
			wantSendCap: 256,
		},
		{
			name: "trusted proxy/uses CF-Connecting-Ip",
			dialHeaders: map[string]string{
				"CF-Connecting-IP": "2.2.2.2",
			},
			trustPeer:   true,
			wantStatus:  http.StatusSwitchingProtocols,
			wantIP:      "2.2.2.2",
			wantSendCap: 256,
		},
		{
			name: "trusted proxy/uses X-Real-Ip",
			dialHeaders: map[string]string{
				"X-Real-Ip": "3.3.3.3",
			},
			trustPeer:   true,
			wantStatus:  http.StatusSwitchingProtocols,
			wantIP:      "3.3.3.3",
			wantSendCap: 256,
		},
		{
			name: "untrusted peer/forwarded headers ignored",
			dialHeaders: map[string]string{
				"CF-Connecting-IP": "2.2.2.2",
				"X-Real-Ip":        "3.3.3.3",
			},
			wantStatus:  http.StatusSwitchingProtocols,
			wantIP:      "127.0.0.1",
			wantSendCap: 256,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hub := NewHub()
			go hub.Run(t.Context())

			middlewareAPI := &API{}
			if tc.trustPeer {
				middlewareAPI.trustedProxies = []netip.Prefix{
					netip.MustParsePrefix("127.0.0.0/8"),
					netip.MustParsePrefix("::1/128"),
				}
			}
			server := httptest.NewServer(middlewareAPI.clientIPMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ServeWs(hub, w, r, nil)
				}),
			))
			t.Cleanup(server.Close)

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			dialHeader := http.Header{}
			for key, value := range tc.dialHeaders {
				dialHeader.Set(key, value)
			}
			clientConn, resp, err := websocket.DefaultDialer.Dial(wsURL, dialHeader)
			if err != nil {
				t.Fatalf("Dialer.Dial(%q) failed: %v", wsURL, err)
			}
			t.Cleanup(func() { _ = clientConn.Close() })

			if got, want := resp.StatusCode, tc.wantStatus; got != want {
				t.Errorf("status code mismatch\ngot:  %d\nwant: %d", got, want)
			}

			registered := waitForHubClient(t, hub)

			fieldTests := []test.FieldAssertion{
				{Name: "ip", Got: registered.ip, Want: tc.wantIP, Mode: test.CompareEqual},
				{Name: "send capacity", Got: cap(registered.send), Want: tc.wantSendCap, Mode: test.CompareEqual},
				{Name: "hub", Got: registered.hub, Want: hub, Mode: test.CompareSamePointer},
				{Name: "conn", Got: registered.conn, Mode: test.CompareNotNil},
			}
			if testErr := test.AssertFields(t, fieldTests, "Client"); testErr != nil {
				t.Fatal(testErr)
			}
		})
	}
}

func TestServeWs__clientAuth(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())
	auth := &clientAuth{
		userID:             "user-a",
		sessionHash:        "hash-1",
		readableServices:   map[string]bool{"svc": true},
		actionableServices: map[string]bool{"svc": true},
		sessionAlive:       func() bool { return true },
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r, auth)
	}))
	t.Cleanup(server.Close)

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dialer.Dial(%q) failed: %v", wsURL, err)
	}
	t.Cleanup(func() { _ = clientConn.Close() })

	registered := waitForHubClient(t, hub)
	fieldTests := []test.FieldAssertion{
		{Name: "userID", Got: registered.userID, Want: auth.userID, Mode: test.CompareEqual},
		{Name: "sessionHash", Got: registered.sessionHash, Want: auth.sessionHash, Mode: test.CompareEqual},
		{Name: `readableServices["svc"]`, Got: registered.readableServices["svc"], Want: true, Mode: test.CompareEqual},
		{Name: `actionableServices["svc"]`, Got: registered.actionableServices["svc"], Want: true, Mode: test.CompareEqual},
		{Name: "sessionAlive", Got: registered.sessionAlive, Mode: test.CompareNotNil},
	}
	if testErr := test.AssertFields(t, fieldTests, "Client"); testErr != nil {
		t.Error(testErr)
	}
}

type wsTestClient struct {
	client *Client
	conn   *websocket.Conn
	server *httptest.Server
	pings  *atomic.Int32
}

func setupWSTestClient(t *testing.T) *wsTestClient {
	t.Helper()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	pings := &atomic.Int32{}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Errorf("Upgrader.Upgrade() failed: %v", err)
					return
				}

				ws.SetPingHandler(func(data string) error {
					pings.Add(1)
					return ws.WriteControl(websocket.PongMessage, []byte(data), time.Now().Add(time.Second))
				})

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

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dialer.Dial(%q) failed: %v", wsURL, err)
	}

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
		pings:  pings,
	}
}

// cleanup closes the [wsTestClient.conn] and the [wsTestClient.server].
func (w *wsTestClient) cleanup() {
	_ = w.conn.Close()
	w.server.Close()
}

// closePeer sends a close frame with code, or drops the [wsTestClient.conn] when code is 0.
func (w *wsTestClient) closePeer(code int) {
	if code == 0 {
		_ = w.conn.Close()
		return
	}

	_ = w.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, ""),
	)
}

func TestServeWs__plain_HTTP(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	ServeWs(hub, rec, req, nil)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Errorf("status code mismatch\ngot:  %d - %q\nwant: %d", got, rec.Body.String(), want)
	}

	time.Sleep(100 * time.Millisecond)

	if registered := hubClientForTest(t, hub); registered != nil {
		t.Errorf("hubClientForTest() mismatch\ngot:  ip=%q\nwant: nil", registered.ip)
	}
}

func TestGetIP(t *testing.T) {
	tests := []struct {
		name       string
		contextIP  *string
		remoteAddr string
		want       string
	}{
		{
			name:       "middleware-resolved IP",
			contextIP:  new("1.1.1.1"),
			remoteAddr: "4.4.4.4:123",
			want:       "1.1.1.1",
		},
		{
			name:       "no middleware. RemoteAddr fallback",
			remoteAddr: "4.4.4.4:123",
			want:       "4.4.4.4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			req.RemoteAddr = tc.remoteAddr
			if tc.contextIP != nil {
				req = req.WithContext(
					context.WithValue(req.Context(), clientIPKey{}, *tc.contextIP))
			}

			got := getIP(req)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestAPI_ForwardedIP(t *testing.T) {
	tests := []struct {
		name    string
		trusted []string
		headers map[string]string
		want    string
	}{
		{
			name: "X-Forwarded-For/single client",
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3",
			},
			want: "3.3.3.3",
		},
		{
			name:    "X-Forwarded-For/rightmost non-proxy is the client",
			trusted: []string{"10.0.0.0/8"},
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3, 10.0.0.5",
			},
			want: "3.3.3.3",
		},
		{
			name:    "X-Forwarded-For/spoofed prefix is ignored",
			trusted: []string{"10.0.0.0/8"},
			headers: map[string]string{
				// Attacker prepends 6.6.6.6; the proxy appends the real 3.3.3.3.
				"X-FORWARDED-FOR": "6.6.6.6, 3.3.3.3, 10.0.0.5",
			},
			want: "3.3.3.3",
		},
		{
			name:    "X-Forwarded-For/malformed entry skipped mid-walk",
			trusted: []string{"10.0.0.0/8"},
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3, not-an-ip, 10.0.0.5",
			},
			want: "3.3.3.3",
		},
		{
			name:    "X-Forwarded-For/multiple distinct trusted proxies",
			trusted: []string{"10.0.0.0/8", "172.16.0.0/12"},
			headers: map[string]string{
				"X-FORWARDED-FOR": "3.3.3.3, 172.16.0.5, 10.0.0.5",
			},
			want: "3.3.3.3",
		},
		{
			name: "X-Forwarded-For/IPv6 client",
			headers: map[string]string{
				"X-FORWARDED-FOR": "2001:db8::1",
			},
			want: "2001:db8::1",
		},
		{
			name: "X-Forwarded-For/takes priority over CF and X-Real-Ip",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-REAL-IP":        "2.2.2.2",
				"X-FORWARDED-FOR":  "3.3.3.3",
			},
			want: "3.3.3.3",
		},
		{
			name:    "X-Forwarded-For/all entries trusted falls back to CF",
			trusted: []string{"10.0.0.0/8"},
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-FORWARDED-FOR":  "10.0.0.1, 10.0.0.2",
			},
			want: "1.1.1.1",
		},
		{
			name: "fallback/CF-Connecting-Ip when no X-Forwarded-For",
			headers: map[string]string{
				"CF-Connecting-IP": "1.1.1.1",
				"X-REAL-IP":        "2.2.2.2",
			},
			want: "1.1.1.1",
		},
		{
			name: "fallback/X-Real-Ip when no X-Forwarded-For or CF",
			headers: map[string]string{
				"X-REAL-IP": "2.2.2.2",
			},
			want: "2.2.2.2",
		},
		{
			name: "no headers",
			want: "",
		},
		{
			name: "invalid header values",
			headers: map[string]string{
				"CF-Connecting-IP": "not-an-ip",
				"X-REAL-IP":        "also-not-an-ip",
				"X-FORWARDED-FOR":  "nope",
			},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			proxies := make([]netip.Prefix, 0, len(tc.trusted))
			for _, p := range tc.trusted {
				pfx, err := netip.ParsePrefix(p)
				if err != nil {
					t.Fatalf("netip.ParsePrefix(%q) failed: %v", p, err)
				}
				proxies = append(proxies, pfx)
			}
			api := &API{trustedProxies: proxies}

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			for header, val := range tc.headers {
				req.Header.Set(header, val)
			}

			got := api.forwardedIP(req)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestRemoteAddrIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{
			name:       "host:port",
			remoteAddr: "4.4.4.4:123",
			want:       "4.4.4.4",
		},
		{
			name:       "bare IP",
			remoteAddr: "5.5.5.5",
			want:       "5.5.5.5",
		},
		{
			name:       "invalid/not an IP",
			remoteAddr: "1111",
			want:       "",
		},
		{
			name:       "invalid/ParseIP fail",
			remoteAddr: "1111:123",
			want:       "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/approvals", nil)
			req.RemoteAddr = tc.remoteAddr

			got := remoteAddrIP(req)

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestClient__hubless(t *testing.T) {
	client := &Client{send: make(chan []byte, 8)}

	// No Hub to leave, so this returns.
	client.unregister()

	if got := client.hubDone(); got != nil {
		t.Error("Client.hubDone() mismatch\ngot:  non-nil\nwant: nil")
	}
}

func TestClient_ReadPump__pongHandler(t *testing.T) {
	tests := []struct {
		name           string
		ping           bool
		wantRegistered bool
	}{
		{
			name:           "pongs extend the read deadline",
			ping:           true,
			wantRegistered: true,
		},
		{
			name:           "no pongs, so the read deadline expires",
			ping:           false,
			wantRegistered: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wsTest := setupWSHubClient(t)
			go wsTest.client.readPump()

			// Runs for three pongWaits, so only pongs can keep the client registered.
			deadline := time.Now().Add(3 * pongWait)
			for time.Now().Before(deadline) {
				if tc.ping {
					err := wsTest.client.conn.WriteControl(
						websocket.PingMessage,
						[]byte{},
						time.Now().Add(time.Second),
					)
					if err != nil {
						t.Fatalf("Conn.WriteControl() failed: %v", err)
					}
				}
				time.Sleep(pongWait / 10)
			}

			if got := wsTest.client.hub.hasClient(wsTest.client); got != tc.wantRegistered {
				t.Errorf("Hub.hasClient() mismatch\ngot:  %t\nwant: %t", got, tc.wantRegistered)
			}
		})
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
			name: "message exceeds read limit (websocket.ErrReadLimit)",
			peerMessages: []string{
				strings.Repeat("a", maxMessageSize+1),
			},
			stdoutRegex: `^$`,
		},
		{
			name:          "expected peer close/normal closure",
			peerCloseCode: websocket.CloseNormalClosure,
			stdoutRegex:   `^$`,
		},
		{
			name:          "expected peer close/going away (tab closed)",
			peerCloseCode: websocket.CloseGoingAway,
			stdoutRegex:   `^$`,
		},
		{
			name:          "expected peer close/no status",
			peerCloseCode: websocket.CloseNoStatusReceived,
			stdoutRegex:   `^$`,
		},
		{
			name:          "unexpected peer close/internal server error",
			peerCloseCode: websocket.CloseInternalServerErr,
			stdoutRegex:   `^ERROR: WebSocket \(127\.0\.0\.1\), websocket: close \d+ .*`,
		},
		{
			name:            "abrupt client close unregisters from hub",
			closeClientConn: true,
			stdoutRegex:     `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			wsTest := setupWSHubClient(t)

			go wsTest.client.readPump()

			for i, msg := range tc.peerMessages {
				if err := wsTest.conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					t.Fatalf("message %d: Conn.WriteMessage() failed: %v", i, err)
				}
			}
			switch {
			case tc.peerCloseCode != 0:
				wsTest.closePeer(tc.peerCloseCode)
			case tc.closeClientConn:
				_ = wsTest.conn.Close()
			}

			waitForClientUnregistered(t, wsTest.client.hub, wsTest.client)

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
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
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			wsTest := setupWSTestClient(t)
			t.Cleanup(wsTest.cleanup)

			if tc.closeConn {
				if err := wsTest.conn.Close(); err != nil {
					t.Fatalf("Conn.Close() failed: %v", err)
				}
			}

			for _, msg := range tc.messages {
				wsTest.client.writeServerMessage([]byte(msg))
			}

			got := readConnMessages(t, wsTest.conn, len(tc.wantMessages))
			assertConnMessages(t, got, tc.wantMessages)

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
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
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			wsTest := setupWSTestClient(t)
			t.Cleanup(wsTest.cleanup)

			for _, msg := range tc.queued {
				wsTest.client.send <- []byte(msg)
			}

			if tc.closeSend {
				close(wsTest.client.send)
			}

			if tc.writeFirst != "" {
				wsTest.client.writeServerMessage([]byte(tc.writeFirst))
			}

			wsTest.client.drainSendMessages()

			time.Sleep(250 * time.Millisecond)
			got := readConnMessages(t, wsTest.conn, len(tc.wantMessages))
			assertConnMessages(t, got, tc.wantMessages)

			if gotLen, wantLen := len(wsTest.client.send), 0; gotLen != wantLen {
				t.Errorf("Client.send length mismatch\ngot:  %d\nwant: %d", gotLen, wantLen)
			}

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}
		})
	}
}

// readConnMessages reads count messages from conn, plus any extra one that arrives shortly after.
// The extra read times out, which leaves conn unreadable.
func readConnMessages(t *testing.T, conn *websocket.Conn, count int) []string {
	t.Helper()

	received := make([]string, 0, count+1)
	for i := range count {
		if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
			t.Fatalf("message %d: Conn.SetReadDeadline() failed: %v", i, err)
		}
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("message %d: Conn.ReadMessage() failed: %v", i, err)
		}
		received = append(received, string(message))
	}

	// A closed conn fails to set the deadline, and holds no extra message.
	if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err == nil {
		if _, message, err := conn.ReadMessage(); err == nil {
			received = append(received, string(message))
		}
	}
	return received
}

func assertConnMessages(t *testing.T, got, want []string) {
	t.Helper()

	if gotLen, wantLen := len(got), len(want); gotLen != wantLen {
		t.Errorf("message count mismatch\ngot:  %d\nwant: %d", gotLen, wantLen)
	}
	for i, wantMsg := range want {
		if i >= len(got) {
			break
		}
		if gotMsg := got[i]; gotMsg != wantMsg {
			t.Errorf("message %d: mismatch\ngot:  %q\nwant: %q", i, gotMsg, wantMsg)
		}
	}
}

// setupWSHubClient returns a registered client with a running hub.
func setupWSHubClient(t *testing.T) *wsTestClient {
	t.Helper()

	wsTest := setupWSTestClient(t)
	t.Cleanup(wsTest.cleanup)
	go wsTest.client.hub.Run(t.Context())
	wsTest.client.hub.register <- wsTest.client
	time.Sleep(100 * time.Millisecond)

	return wsTest
}

// waitForHubClient returns the hub's one client, failing unless it registers within 1s.
func waitForHubClient(t *testing.T, hub *Hub) *Client {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if registered := hubClientForTest(t, hub); registered != nil {
			return registered
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("hubClientForTest() mismatch\ngot:  nil\nwant: non-nil")
	return nil
}

// waitForClientUnregistered fails unless client leaves hub within 2s.
func waitForClientUnregistered(t *testing.T, hub *Hub, client *Client) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !hub.hasClient(client) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("Hub.hasClient() mismatch\ngot:  true after 2s\nwant: false")
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
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			wsTest := setupWSHubClient(t)

			go wsTest.client.writePump()

			for _, msg := range tc.messages {
				wsTest.client.send <- []byte(msg)
			}

			time.Sleep(250 * time.Millisecond)
			got := readConnMessages(t, wsTest.conn, len(tc.wantMessages))
			assertConnMessages(t, got, tc.wantMessages)

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
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
		sessionDead         bool
		wantExit            bool
		wantPings           bool
		stdoutRegex         string
	}{
		{
			name:        "closes on empty send channel",
			closeSend:   true,
			wantExit:    true,
			stdoutRegex: `VERBOSE: .*Closing the connection \(writePump\)`,
		},
		{
			name:        "session death closes the connection",
			sessionDead: true,
			wantExit:    true,
			stdoutRegex: `VERBOSE: .*Closing the connection \(session expired\)`,
		},
		{
			name: "logs write failure when connection closed",
			messages: []string{
				`{"version":null,"type":"VERSION","page":"home"}`,
			},
			closeConnBeforePump: true,
			wantExit:            true,
			stdoutRegex:         `^ERROR:.*Writing JSON to the WebSocket failed for VERSION`,
		},
		{
			name:        "sends ping frames",
			wantExit:    false,
			wantPings:   true,
			stdoutRegex: `^$`,
		},
		{
			name:               "ping write failure exits writePump",
			closeConnAfterPump: true,
			wantExit:           true,
			stdoutRegex:        `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Serial: captures stdout.

			releaseStdout := test.CaptureLog(t, logx.Default())

			wsTest := setupWSHubClient(t)

			if tc.sessionDead {
				wsTest.client.sessionAlive = func() bool { return false }
			}

			if tc.closeConnBeforePump {
				if err := wsTest.conn.Close(); err != nil {
					t.Fatalf("before pump: Conn.Close() failed: %v", err)
				}
			}

			exited := make(chan struct{})
			startPump := func() {
				go func() {
					wsTest.client.writePump()
					close(exited)
				}()
			}

			for _, msg := range tc.messages {
				wsTest.client.send <- []byte(msg)
			}

			if tc.closeConnBeforePump {
				startPump()
			} else {
				startPump()
				if tc.closeSend {
					close(wsTest.client.send)
				}

				if tc.closeConnAfterPump {
					if err := wsTest.conn.Close(); err != nil {
						t.Fatalf("after pump: Conn.Close() failed: %v", err)
					}
				}
			}

			time.Sleep(300 * time.Millisecond)
			gotExit := false
			select {
			case <-exited:
				gotExit = true
			default:
			}
			if gotExit != tc.wantExit {
				t.Errorf(
					"Client.writePump() exited mismatch\ngot:  %t\nwant: %t",
					gotExit, tc.wantExit,
				)
			}

			if tc.wantPings {
				if got := wsTest.pings.Load(); got == 0 {
					t.Errorf("wsTestClient.pings mismatch\ngot:  %d\nwant: >0", got)
				}
			}

			stdout := releaseStdout()
			if !util.RegexCheck(tc.stdoutRegex, stdout) {
				t.Errorf("stdout mismatch\ngot:  %q\nwant: %q", stdout, tc.stdoutRegex)
			}
		})
	}
}

// hubClientForTest returns hub's only Client, nil if it has none, failing if it has more.
func hubClientForTest(t *testing.T, hub *Hub) *Client {
	t.Helper()

	clients := hub.clientList()
	if len(clients) == 0 {
		return nil
	}

	if len(clients) > 1 {
		t.Fatalf("Hub.clientList() length mismatch\ngot:  %d\nwant: 1", len(clients))
	}

	return clients[0]
}
