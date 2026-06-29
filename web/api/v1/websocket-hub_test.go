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
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
)

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
				t.Errorf("%s client %q failed to unmarshal broadcast: %v", prefix, name, unmarshalErr)
				continue
			}
			if gotMsg != msg {
				t.Errorf(
					"%s client %q message mismatch\ngot:  %v\nwant: %v",
					prefix, name, gotMsg, msg,
				)
			}
		default:
			t.Errorf("%s client %q did not receive the broadcast", prefix, name)
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
