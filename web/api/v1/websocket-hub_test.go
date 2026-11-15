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
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	fieldTests := []test.FieldAssertion{
		{Name: "Broadcast", Got: hub.Broadcast, Mode: test.CompareNotNil},
		{Name: "register", Got: hub.register, Mode: test.CompareNotNil},
		{Name: "unregister", Got: hub.unregister, Mode: test.CompareNotNil},
		{Name: "query", Got: hub.query, Mode: test.CompareNotNil},
		{Name: "clients", Got: hub.clients, Mode: test.CompareNotNil},
	}
	if testErr := test.AssertFields(t, fieldTests, "Hub"); testErr != nil {
		t.Fatal(testErr)
	}
}

// hubQuery runs fn against the clients map on the Run goroutine, so a live hub can
// be read race-free. Panics if Run() does not answer within 5s.
func hubQuery[T any](h *Hub, fn func(map[*Client]bool) T) T {
	timeout := time.After(5 * time.Second)
	done := make(chan T, 1)
	select {
	case h.query <- func(clients map[*Client]bool) { done <- fn(clients) }:
	case <-timeout:
		panic("hubQuery: Hub.Run() did not accept the query within 5s")
	}

	select {
	case got := <-done:
		return got
	case <-timeout:
		panic("hubQuery: Hub.Run() did not answer the query within 5s")
	}
}

// drainBroadcasts waits for [Hub.Run] to dequeue every queued broadcast, then
// syncs with its goroutine so the last broadcast() has finished sending.
func drainBroadcasts(t *testing.T, h *Hub) {
	t.Helper()

	for range 200 {
		if len(h.Broadcast) == 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	hubQuery(h, func(map[*Client]bool) bool { return true })
}

// hasClient reports whether client is registered (via the Run loop, so safe on a live hub).
func (h *Hub) hasClient(client *Client) bool {
	return hubQuery(
		h,
		func(clients map[*Client]bool) bool {
			return clients[client]
		},
	)
}

// clientList snapshots the registered clients (via the Run loop, so safe on a live hub).
func (h *Hub) clientList() []*Client {
	return hubQuery(
		h,
		func(clients map[*Client]bool) []*Client {
			list := make([]*Client, 0, len(clients))
			for client := range clients {
				list = append(list, client)
			}
			return list
		},
	)
}

func TestAnnounceMSG_ServiceID(t *testing.T) {
	tests := []struct {
		name string
		msg  AnnounceMSG
		want string
	}{
		{
			name: "VERSION-style message carries service_data.id",
			msg: AnnounceMSG{
				Type: "VERSION", SubType: "QUERY",
				ServiceData: &struct {
					ID string `json:"id"`
				}{
					ID: "argus",
				},
			},
			want: "argus",
		},
		{
			name: "DELETE message carries the ID as its sub-type",
			msg:  AnnounceMSG{Type: "DELETE", SubType: "argus"},
			want: "argus",
		},
		{
			name: "EDIT message carries the ID as its sub-type (unchanged ID stripped from service_data)",
			msg: AnnounceMSG{
				Type: "EDIT", SubType: "argus",
				ServiceData: &struct {
					ID string `json:"id"`
				}{
					ID: "",
				},
			},
			want: "argus",
		},
		{
			name: "ORDER message concerns no single service",
			msg:  AnnounceMSG{Type: "SERVICE", SubType: "ORDER"},
			want: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.msg.serviceID()

			if got != tc.want {
				t.Errorf("mismatch\ngot:  %q\nwant: %q", got, tc.want)
			}
		})
	}
}

func TestHub_AddClient(t *testing.T) {
	hub := NewHub()
	client := testClient()
	otherClient := testClient()

	hub.addClient(&client)
	hub.addClient(&otherClient)

	if !hub.clients[&client] {
		t.Error("Hub.clients[client] mismatch\ngot:  false\nwant: true")
	}
	if !hub.clients[&otherClient] {
		t.Error("Hub.clients[otherClient] mismatch\ngot:  false\nwant: true")
	}
}

func TestHub_RemoveClient(t *testing.T) {
	hub := NewHub()
	client := testClient()
	otherClient := testClient()
	hub.addClient(&client)
	hub.addClient(&otherClient)
	if !hub.clients[&client] || !hub.clients[&otherClient] {
		t.Fatalf(
			"before removal: Hub.clients mismatch\n"+
				"got:  client=%t, otherClient=%t\nwant: client=true, otherClient=true",
			hub.clients[&client], hub.clients[&otherClient],
		)
	}

	hub.removeClient(&client)
	hub.removeClient(&otherClient)

	if hub.clients[&client] {
		t.Error("after removal: Hub.clients[client] mismatch\ngot:  true\nwant: false")
	}
	if hub.clients[&otherClient] {
		t.Error("after removal: Hub.clients[otherClient] mismatch\ngot:  true\nwant: false")
	}
}

func TestHub_Broadcast(t *testing.T) {
	client := testClient()
	hub := client.hub
	hub.addClient(&client)

	msg := AnnounceMSG{
		Type:    "test",
		SubType: "something",
	}

	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf("decode.Marshal() failed: %v", err)
	}
	hub.broadcast(data)

	var got []byte
	select {
	case got = <-client.send:
	case <-time.After(time.Second):
		t.Fatal("Client.send message count mismatch\ngot:  0\nwant: 1")
	}

	var gotMsg AnnounceMSG
	if err := decode.Unmarshal("json", got, &gotMsg); err != nil {
		t.Fatalf("decode.Unmarshal() failed: %v", err)
	}
	if gotMsg != msg {
		t.Errorf("Client.send message mismatch\ngot:  %q\nwant: %q", got, data)
	}
}

func TestHub_Broadcast_allClients(t *testing.T) {
	hub := NewHub()
	clientA := testClient()
	clientB := testClient()
	clientC := testClient()
	hub.addClient(&clientA)
	hub.addClient(&clientB)
	hub.addClient(&clientC)

	msg := AnnounceMSG{
		Type:    "test",
		SubType: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf("decode.Marshal() failed: %v", err)
	}

	hub.broadcast(data)

	for _, c := range []struct {
		name   string
		client *Client
	}{
		{name: "A", client: &clientA},
		{name: "B", client: &clientB},
		{name: "C", client: &clientC},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			select {
			case got := <-c.client.send:
				var gotMsg AnnounceMSG
				if unmarshalErr := decode.Unmarshal("json", got, &gotMsg); unmarshalErr != nil {
					t.Fatalf("decode.Unmarshal() failed: %v", unmarshalErr)
				}
				if gotMsg != msg {
					t.Errorf("Client.send message mismatch\ngot:  %q\nwant: %q", got, data)
				}
			default:
				t.Error("Client.send message count mismatch\ngot:  0\nwant: 1")
			}
		})
	}
}

func TestHub_Broadcast__dropsFullClient(t *testing.T) {
	hub := NewHub()
	slowClient := Client{
		ip:   "1.1.1.1",
		send: make(chan []byte, 1),
	}
	staleMessage := `{"type":"test"}`
	slowClient.send <- []byte(staleMessage)
	readyClient := testClient()
	hub.addClient(&slowClient)
	hub.addClient(&readyClient)

	msg := AnnounceMSG{
		Type:    "test",
		SubType: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf("decode.Marshal() failed: %v", err)
	}

	hub.broadcast(data)

	if hub.clients[&slowClient] {
		t.Error("slow client: Hub.clients[slowClient] mismatch\ngot:  true\nwant: false (dropped)")
	}

	select {
	case got, ok := <-slowClient.send:
		if !ok || string(got) != staleMessage {
			t.Errorf(
				"slow client: first Client.send receive mismatch\ngot:  %q (ok=%t)\nwant: %q",
				got, ok, staleMessage,
			)
		}
	case <-time.After(time.Second):
		t.Errorf(
			"slow client: first Client.send receive mismatch\ngot:  empty\nwant: %q",
			staleMessage,
		)
	}
	select {
	case _, ok := <-slowClient.send:
		if ok {
			t.Error("slow client: second Client.send receive mismatch\ngot:  message\nwant: closed")
		}
	case <-time.After(time.Second):
		t.Error("slow client: second Client.send receive mismatch\ngot:  open\nwant: closed")
	}

	if !hub.clients[&readyClient] {
		t.Error("ready client: Hub.clients[readyClient] mismatch\ngot:  false\nwant: true")
	}

	var got []byte
	select {
	case got = <-readyClient.send:
	case <-time.After(time.Second):
		t.Fatal("ready client: Client.send message count mismatch\ngot:  0\nwant: 1")
	}
	var gotMsg AnnounceMSG
	if unmarshalErr := decode.Unmarshal("json", got, &gotMsg); unmarshalErr != nil {
		t.Fatalf("ready client: decode.Unmarshal() failed: %v", unmarshalErr)
	}
	if gotMsg != msg {
		t.Errorf("ready client: Client.send message mismatch\ngot:  %q\nwant: %q", got, data)
	}
}

func TestHub_Broadcast__invalid(t *testing.T) {
	client := testClient()
	hub := client.hub
	hub.addClient(&client)

	hub.broadcast([]byte("key: value\nkey: value"))

	got := len(client.send)
	want := 0
	if got != want {
		t.Errorf("Client.send message count mismatch\ngot:  %d\nwant: %d", got, want)
	}
}

func TestHub_Broadcast__readableFiltering(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())

	unrestricted := &Client{
		hub:  hub,
		send: make(chan []byte, 8),
	}
	restricted := &Client{
		hub:              hub,
		send:             make(chan []byte, 8),
		readableServices: map[string]bool{"allowed-svc": true},
	}
	hub.register <- unrestricted
	hub.register <- restricted

	messages := [][]byte{
		[]byte(`{"page":"APPROVALS","type":"VERSION","sub_type":"QUERY","service_data":{"id":"allowed-svc"}}`),
		[]byte(`{"page":"APPROVALS","type":"VERSION","sub_type":"QUERY","service_data":{"id":"secret-svc"}}`),
		[]byte(`{"page":"APPROVALS","type":"SERVICE","sub_type":"ORDER","order":["allowed-svc","secret-svc"]}`),
		[]byte(`{"page":"APPROVALS","type":"DELETE","sub_type":"secret-svc"}`),
	}
	for _, message := range messages {
		hub.Broadcast <- message
	}
	drainBroadcasts(t, hub)

	if got, want := len(unrestricted.send), len(messages); got != want {
		t.Errorf(
			"unrestricted client: Client.send message count mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}

	// Only allowed-svc's message: ORDER names no service.
	if got, want := len(restricted.send), 1; got != want {
		t.Fatalf(
			"restricted client: Client.send message count mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}
	received := <-restricted.send
	if want := "allowed-svc"; !strings.Contains(string(received), want) {
		t.Errorf(
			"restricted client: Client.send message mismatch\ngot:  %q\nwant: %q",
			received, want,
		)
	}
}

func TestHub_Broadcast__actionableFiltering(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())

	unrestricted := &Client{
		hub:  hub,
		send: make(chan []byte, 8),
	}
	readOnly := &Client{
		hub:                hub,
		send:               make(chan []byte, 8),
		readableServices:   map[string]bool{"svc": true},
		actionableServices: map[string]bool{},
	}
	actioner := &Client{
		hub:                hub,
		send:               make(chan []byte, 8),
		readableServices:   map[string]bool{"svc": true},
		actionableServices: map[string]bool{"svc": true},
	}
	for _, client := range []*Client{unrestricted, readOnly, actioner} {
		hub.register <- client
	}

	messages := [][]byte{
		[]byte(`{"page":"APPROVALS","type":"VERSION","sub_type":"QUERY","service_data":{"id":"svc"}}`),
		[]byte(`{"page":"APPROVALS","type":"COMMAND","sub_type":"EVENT","service_data":{"id":"svc"},"command_data":{"bash -c true":{"failed":false}}}`),
		[]byte(`{"page":"APPROVALS","type":"WEBHOOK","sub_type":"EVENT","service_data":{"id":"svc"},"webhook_data":{"wh":{"failed":false}}}`),
	}
	for _, message := range messages {
		hub.Broadcast <- message
	}
	drainBroadcasts(t, hub)

	for _, tc := range []struct {
		name   string
		client *Client
	}{
		{name: "unrestricted", client: unrestricted},
		{name: "actioner", client: actioner},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := len(tc.client.send), len(messages); got != want {
				t.Errorf("Client.send message count mismatch\ngot:  %d\nwant: %d", got, want)
			}
		})
	}

	// An empty action set withholds COMMAND and WEBHOOK.
	if got, want := len(readOnly.send), 1; got != want {
		t.Fatalf(
			"read-only client: Client.send message count mismatch\ngot:  %d\nwant: %d",
			got, want,
		)
	}
	received := <-readOnly.send
	if want := `"type":"VERSION"`; !strings.Contains(string(received), want) {
		t.Errorf(
			"read-only client: Client.send message mismatch\ngot:  %q\nwant: %q",
			received, want,
		)
	}
}

func TestNonEmptySet(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   map[string]bool
	}{
		{
			name:   "empty/nil",
			values: nil,
			want:   map[string]bool{},
		},
		{
			name:   "empty/no values",
			values: []string{},
			want:   map[string]bool{},
		},
		{
			name: "empty/only blanks",
			// A blank value is the identity a client has when auth is disabled.
			values: []string{"", "", ""},
			want:   map[string]bool{},
		},
		{
			name:   "collects/single value",
			values: []string{"user-a"},
			want:   map[string]bool{"user-a": true},
		},
		{
			name:   "collects/several values",
			values: []string{"user-a", "user-b"},
			want:   map[string]bool{"user-a": true, "user-b": true},
		},
		{
			name:   "collects/drops the blanks between them",
			values: []string{"", "user-a", "", "user-b", ""},
			want:   map[string]bool{"user-a": true, "user-b": true},
		},
		{
			name:   "collects/deduplicates repeats",
			values: []string{"user-a", "user-a", "user-a"},
			want:   map[string]bool{"user-a": true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := nonEmptySet(tc.values)

			if testErr := test.AssertMapEqual(t, got, tc.want, "set"); testErr != nil {
				t.Error(testErr)
			}
		})
	}
}

// assertKicked checks which clients were kicked (send channel closed and
// deregistered) and which were spared, after a Kick*Clients call.
func assertKicked(t *testing.T, hub *Hub, kicked, spared []*Client) {
	t.Helper()

	for i, client := range kicked {
		if hub.hasClient(client) {
			t.Errorf("kicked %d: Hub.hasClient() mismatch\ngot:  true\nwant: false", i)
		}
		select {
		case _, open := <-client.send:
			if open {
				t.Errorf("kicked %d: Client.send mismatch\ngot:  message\nwant: closed", i)
			}
		case <-time.After(time.Second):
			t.Errorf("kicked %d: Client.send mismatch\ngot:  open\nwant: closed", i)
		}
	}
	for i, client := range spared {
		if !hub.hasClient(client) {
			t.Errorf("spared %d: Hub.hasClient() mismatch\ngot:  false\nwant: true", i)
		}
	}
}

func TestHub_KickUserClients(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())
	targetTab1 := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a"}
	targetTab2 := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a"}
	otherUser := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-b"}
	anonymous := &Client{hub: hub, send: make(chan []byte, 8)}
	for _, client := range []*Client{targetTab1, targetTab2, otherUser, anonymous} {
		hub.register <- client
	}

	// The empty ID must not match the anonymous client.
	hub.KickUserClients("user-a", "")

	assertKicked(t, hub,
		[]*Client{targetTab1, targetTab2},
		[]*Client{otherUser, anonymous},
	)

	hub.KickUserClients()
	hub.KickUserClients("")
	assertKicked(t, hub,
		nil,
		[]*Client{otherUser, anonymous},
	)
}

func TestHub_KickSessionClients(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())
	target := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a", sessionHash: "hash-1"}
	sameUserOtherSession := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a", sessionHash: "hash-2"}
	otherUser := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-b", sessionHash: "hash-3"}
	anonymous := &Client{hub: hub, send: make(chan []byte, 8)}
	for _, client := range []*Client{target, sameUserOtherSession, otherUser, anonymous} {
		hub.register <- client
	}

	// The empty hash must not match the anonymous client.
	hub.KickSessionClients("hash-1", "")

	assertKicked(t, hub,
		[]*Client{target},
		[]*Client{sameUserOtherSession, otherUser, anonymous},
	)

	hub.KickSessionClients()
	hub.KickSessionClients("")
	assertKicked(t, hub,
		nil,
		[]*Client{sameUserOtherSession, otherUser, anonymous},
	)
}

func TestHub_KickRestrictedClients(t *testing.T) {
	hub := NewHub()
	go hub.Run(t.Context())
	unrestricted := &Client{hub: hub, send: make(chan []byte, 8)}
	restricted := &Client{hub: hub, send: make(chan []byte, 8), readableServices: map[string]bool{"svc": true}}
	restrictedToNothing := &Client{hub: hub, send: make(chan []byte, 8), readableServices: map[string]bool{}}
	actionRestricted := &Client{hub: hub, send: make(chan []byte, 8), actionableServices: map[string]bool{"svc": true}}
	for _, client := range []*Client{unrestricted, restricted, restrictedToNothing, actionRestricted} {
		hub.register <- client
	}

	hub.KickRestrictedClients()

	// Any non-nil set restricts, even an empty one.
	assertKicked(t, hub,
		[]*Client{restricted, restrictedToNothing, actionRestricted},
		[]*Client{unrestricted},
	)
}

func TestHub_Run__stopsWithContext(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(t.Context())
	stopped := make(chan struct{})
	go func() {
		hub.Run(ctx)
		close(stopped)
	}()
	client := &Client{hub: hub, send: make(chan []byte, 8)}
	hub.register <- client

	cancel()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("Hub.Run() mismatch\ngot:  running after 1s\nwant: returned")
	}

	unregistered := make(chan struct{})
	go func() {
		client.unregister()
		close(unregistered)
	}()
	select {
	case <-unregistered:
	case <-time.After(time.Second):
		t.Error("Client.unregister() mismatch\ngot:  blocked after 1s\nwant: returned")
	}

	kicked := make(chan struct{})
	go func() {
		hub.KickUserClients("user-a")
		close(kicked)
	}()
	select {
	case <-kicked:
	case <-time.After(time.Second):
		t.Error("Hub.KickUserClients() mismatch\ngot:  blocked after 1s\nwant: returned")
	}
}
