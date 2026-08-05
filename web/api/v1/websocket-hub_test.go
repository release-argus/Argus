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
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

func TestNewHub(t *testing.T) {
	// GIVEN: we want a WebSocket Hub.

	// WHEN: we create a new one with NewHub.
	hub := NewHub()

	prefix := fmt.Sprintf("%s\nNewHub()", packageName)

	// THEN: it returns a Hub with all the channels and maps initialised.
	fieldTests := []test.FieldAssertion{
		{Name: "Broadcast", Got: hub.Broadcast, Want: nil, Mode: test.CompareNotEqual},
		{Name: "register", Got: hub.register, Want: nil, Mode: test.CompareNotEqual},
		{Name: "unregister", Got: hub.unregister, Want: nil, Mode: test.CompareNotEqual},
		{Name: "query", Got: hub.query, Want: nil, Mode: test.CompareNotEqual},
		{Name: "clients", Got: hub.clients, Want: nil, Mode: test.CompareNotEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}
}

// hubQuery runs fn against the clients map on the Run goroutine, so a live hub can
// be read race-free. Requires Run() to be active.
func hubQuery[T any](h *Hub, fn func(map[*Client]bool) T) T {
	done := make(chan T, 1)
	h.query <- func(clients map[*Client]bool) {
		done <- fn(clients)
	}
	return <-done
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
	// GIVEN: announce messages of every shape.
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
		{
			name: "top-level service_id is honoured",
			msg:  AnnounceMSG{Type: "OTHER", ServiceID: "argus"},
			want: "argus",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prefix := fmt.Sprintf("%s\nAnnounceMSG.serviceID()", packageName)

			// WHEN: the service ID is extracted.
			got := tc.msg.serviceID()

			// THEN: it matches expectations.
			if got != tc.want {
				t.Errorf(
					"%s\nresult mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.want,
				)
			}
		})
	}
}

func TestHub_AddClient(t *testing.T) {
	// GIVEN: a Hub and two clients (not connected).
	hub := NewHub()
	client := testClient()
	otherClient := testClient()

	// WHEN: the clients register.
	hub.addClient(&client)
	hub.addClient(&otherClient)

	// THEN: those client are both registered to the Hub.
	if !hub.clients[&client] {
		t.Errorf("%s\nclient 1 wasn't registered to the Hub with addClient", packageName)
	}
	if !hub.clients[&otherClient] {
		t.Errorf("%s\nclient 2 wasn't registered to the Hub with addClient", packageName)
	}
}

func TestHub_RemoveClient(t *testing.T) {
	// GIVEN: a Hub with two registered clients.
	hub := NewHub()
	client := testClient()
	otherClient := testClient()
	hub.addClient(&client)
	hub.addClient(&otherClient)
	if !hub.clients[&client] || !hub.clients[&otherClient] {
		t.Errorf("%s\nclient wasn't registered to the Hub", packageName)
	}

	// WHEN: the clients disconnect.
	hub.removeClient(&client)
	hub.removeClient(&otherClient)

	// THEN: those client are unregistered from the Hub.
	if hub.clients[&client] {
		t.Errorf(
			"%s\nclient 1 should have been removed from the Hub after removeClient\nremaining clients: %d",
			packageName, len(hub.clients),
		)
	}
	if hub.clients[&otherClient] {
		t.Errorf(
			"%s\nclient 2 should have been removed from the Hub after removeClient\nremaining clients: %d",
			packageName, len(hub.clients),
		)
	}
}

func TestHub_Broadcast(t *testing.T) {
	prefix := fmt.Sprintf("%s\nHub.broadcast()", packageName)

	// GIVEN: a Hub with a registered client.
	client := testClient()
	hub := client.hub
	hub.addClient(&client)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}

	// WHEN: that message is broadcast.
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf(
			"%s failed to marshal broadcast message: %v",
			prefix, err,
		)
	}
	hub.broadcast(data)

	// THEN: that message is broadcast to the client.
	got := <-client.send
	var gotMsg AnnounceMSG
	_ = decode.Unmarshal("json", got, &gotMsg)
	if gotMsg != msg {
		t.Errorf(
			"%s message should have been received by the client channel\ngot:  %v\nwant: %v",
			prefix, gotMsg, msg,
		)
	}
}

func TestHub_Broadcast_allClients(t *testing.T) {
	prefix := fmt.Sprintf("%s\nHub.broadcast()", packageName)

	// GIVEN: a Hub with multiple registered clients.
	hub := NewHub()
	clientA := testClient()
	clientB := testClient()
	clientC := testClient()
	hub.addClient(&clientA)
	hub.addClient(&clientB)
	hub.addClient(&clientC)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf(
			"%s failed to marshal broadcast message: %v",
			prefix, err,
		)
	}

	// WHEN: that message is broadcast.
	hub.broadcast(data)

	// THEN: every registered client receives the message.
	for name, client := range map[string]*Client{
		"A": &clientA,
		"B": &clientB,
		"C": &clientC,
	} {
		select {
		case got := <-client.send:
			var gotMsg AnnounceMSG
			if unmarshalErr := decode.Unmarshal("json", got, &gotMsg); unmarshalErr != nil {
				t.Errorf(
					"%s client %q failed to unmarshal broadcast: %v",
					prefix, name, unmarshalErr,
				)
				continue
			}
			if gotMsg != msg {
				t.Errorf(
					"%s client %q message mismatch\ngot:  %v\nwant: %v",
					prefix, name, gotMsg, msg,
				)
			}
		default:
			t.Errorf(
				"%s client %q did not receive the broadcast",
				prefix, name,
			)
		}
	}
}

func TestHub_Broadcast__dropsFullClient(t *testing.T) {
	prefix := fmt.Sprintf("%s\nHub.broadcast()", packageName)

	// GIVEN: a hub with a client whose outbound buffer is full and another with capacity.
	hub := NewHub()
	slowClient := Client{
		ip:   "1.1.1.1",
		send: make(chan []byte, 1),
	}
	slowClient.send <- []byte(`{"type":"test"}`)
	readyClient := testClient()
	hub.addClient(&slowClient)
	hub.addClient(&readyClient)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf(
			"%s failed to marshal broadcast message: %v",
			prefix, err,
		)
	}

	// WHEN: that message is broadcast.
	hub.broadcast(data)

	// THEN: the slow client is removed from the Hub.
	if hub.clients[&slowClient] {
		t.Errorf("%s slow client should have been removed from the Hub as outbound buffer is full", prefix)
	}

	// AND: its send channel is closed.
	if _, ok := <-slowClient.send; !ok {
		t.Errorf("%s slow client send channel should still hold the previous message", prefix)
	}
	if _, ok := <-slowClient.send; ok {
		t.Errorf("%s slow client send channel should be empty as it was closed", prefix)
	}

	// AND: the ready client is still registered.
	if !hub.clients[&readyClient] {
		t.Errorf("%s ready client should still be registered", prefix)
	}

	// AND: it received the message.
	got := <-readyClient.send
	var gotMsg AnnounceMSG
	if unmarshalErr := decode.Unmarshal("json", got, &gotMsg); unmarshalErr != nil {
		t.Fatalf(
			"%s ready client failed to unmarshal broadcast: %v",
			prefix, unmarshalErr,
		)
	}
	if gotMsg != msg {
		t.Errorf(
			"%s ready client message mismatch\ngot:  %v\nwant: %v",
			prefix, gotMsg, msg,
		)
	}
}

func TestHub_Broadcast__invalid(t *testing.T) {
	// GIVEN: a Hub with a registered Client.
	client := testClient()
	hub := client.hub
	hub.addClient(&client)

	// AND: an invalid message.
	msg := []byte("key: value\nkey: value")

	// WHEN: that message is broadcast.
	data, _ := decode.Marshal("json", msg)
	hub.broadcast(data)

	// THEN: that message is NOT sent to the client.
	got := len(client.send)
	want := 0
	if got != want {
		t.Errorf(
			"%s\nHub.broadcast() message should have failed Unmarshal and not been sent\n"+
				"got:  %d\nwant: %d",
			packageName, got, want,
		)
	}
}

func TestHub_Broadcast__filtering(t *testing.T) {
	// GIVEN: a running hub with an unrestricted and a restricted client.
	hub := NewHub()
	go hub.Run()

	unrestricted := &Client{
		hub:  hub,
		send: make(chan []byte, 8),
	}
	restricted := &Client{
		hub:             hub,
		send:            make(chan []byte, 8),
		allowedServices: map[string]bool{"allowed-svc": true},
	}
	hub.register <- unrestricted
	hub.register <- restricted

	prefix := fmt.Sprintf("%s\nHub broadcast filtering", packageName)

	// WHEN: messages about different services are broadcast.
	messages := [][]byte{
		[]byte(`{"page":"APPROVALS","type":"VERSION","sub_type":"QUERY","service_data":{"id":"allowed-svc"}}`),
		[]byte(`{"page":"APPROVALS","type":"VERSION","sub_type":"QUERY","service_data":{"id":"secret-svc"}}`),
		[]byte(`{"page":"APPROVALS","type":"SERVICE","sub_type":"ORDER","order":["allowed-svc","secret-svc"]}`),
		[]byte(`{"page":"APPROVALS","type":"DELETE","sub_type":"secret-svc"}`),
	}
	for _, message := range messages {
		hub.Broadcast <- message
	}
	// Wait for the hub to work through the (buffered) broadcasts.
	for i := 0; i < 200 && len(unrestricted.send) < len(messages); i++ {
		time.Sleep(5 * time.Millisecond)
	}

	// THEN: the unrestricted client received everything.
	if got := len(unrestricted.send); got != len(messages) {
		t.Errorf(
			"%s\nunrestricted message count mismatch\ngot:  %d\nwant: %d",
			prefix, got, len(messages),
		)
	}

	// AND: the restricted client received only its service's message.
	if got := len(restricted.send); got != 1 {
		t.Fatalf(
			"%s\nrestricted message count mismatch\ngot:  %d\nwant: 1",
			prefix, got,
		)
	}
	received := <-restricted.send
	if !strings.Contains(string(received), "allowed-svc") {
		t.Errorf(
			"%s\nrestricted client received the wrong message: %s",
			prefix, received,
		)
	}
}

// assertKicked checks which clients were kicked (send channel closed and
// deregistered) and which were spared, after a Kick*Clients call.
func assertKicked(t *testing.T, hub *Hub, prefix string, kicked, spared []*Client) {
	t.Helper()

	for i, client := range kicked {
		if hub.hasClient(client) {
			t.Errorf(
				"%s\nkicked client %d should be deregistered",
				prefix, i,
			)
		}
		if _, open := <-client.send; open {
			t.Errorf(
				"%s\nkicked client %d's send channel should be closed",
				prefix, i,
			)
		}
	}
	for i, client := range spared {
		if !hub.hasClient(client) {
			t.Errorf(
				"%s\nspared client %d should stay registered",
				prefix, i,
			)
		}
	}
}

func TestHub_KickUserClients(t *testing.T) {
	// GIVEN: a running hub with clients of several users,
	// plus one with no user identity (auth disabled).
	hub := NewHub()
	go hub.Run()
	targetTab1 := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a"}
	targetTab2 := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a"}
	otherUser := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-b"}
	anonymous := &Client{hub: hub, send: make(chan []byte, 8)}
	for _, client := range []*Client{targetTab1, targetTab2, otherUser, anonymous} {
		hub.register <- client
	}

	prefix := fmt.Sprintf("%s\nHub.KickUserClients()", packageName)

	// WHEN: one user's clients are kicked
	// (an empty ID must never match the anonymous client).
	hub.KickUserClients("user-a", "")

	// THEN: only that user's clients are kicked; the rest are spared.
	assertKicked(t, hub, prefix,
		[]*Client{targetTab1, targetTab2},
		[]*Client{otherUser, anonymous},
	)

	// AND: kicking no users is a no-op.
	hub.KickUserClients()
	hub.KickUserClients("")
	assertKicked(t, hub, prefix,
		nil,
		[]*Client{otherUser, anonymous},
	)
}

func TestHub_KickSessionClients(t *testing.T) {
	// GIVEN: a running hub with clients of several sessions,
	// plus one with no session (auth disabled).
	hub := NewHub()
	go hub.Run()
	target := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a", sessionHash: "hash-1"}
	sameUserOtherSession := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-a", sessionHash: "hash-2"}
	otherUser := &Client{hub: hub, send: make(chan []byte, 8), userID: "user-b", sessionHash: "hash-3"}
	anonymous := &Client{hub: hub, send: make(chan []byte, 8)}
	for _, client := range []*Client{target, sameUserOtherSession, otherUser, anonymous} {
		hub.register <- client
	}

	prefix := fmt.Sprintf("%s\nHub.KickSessionClients()", packageName)

	// WHEN: one session's clients are kicked
	// (an empty hash must never match the anonymous client).
	hub.KickSessionClients("hash-1", "")

	// THEN: only that session's client is kicked - even the same user's
	// other session survives.
	assertKicked(t, hub, prefix,
		[]*Client{target},
		[]*Client{sameUserOtherSession, otherUser, anonymous},
	)

	// AND: kicking no sessions is a no-op.
	hub.KickSessionClients()
	hub.KickSessionClients("")
	assertKicked(t, hub, prefix,
		nil,
		[]*Client{sameUserOtherSession, otherUser, anonymous},
	)
}

func TestHub_KickRestrictedClients(t *testing.T) {
	// GIVEN: a running hub with restricted and unrestricted clients.
	hub := NewHub()
	go hub.Run()
	unrestricted := &Client{hub: hub, send: make(chan []byte, 8)}
	restricted := &Client{hub: hub, send: make(chan []byte, 8), allowedServices: map[string]bool{"svc": true}}
	restrictedToNothing := &Client{hub: hub, send: make(chan []byte, 8), allowedServices: map[string]bool{}}
	for _, client := range []*Client{unrestricted, restricted, restrictedToNothing} {
		hub.register <- client
	}

	prefix := fmt.Sprintf("%s\nHub.KickRestrictedClients()", packageName)

	// WHEN: the restricted clients are kicked.
	hub.KickRestrictedClients()

	// THEN: only clients with a permitted-service set are kicked.
	assertKicked(t, hub, prefix,
		[]*Client{restricted, restrictedToNothing},
		[]*Client{unrestricted},
	)
}
