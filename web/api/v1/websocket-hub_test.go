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
		{Name: "clients", Got: hub.clients, Want: nil, Mode: test.CompareNotEqual},
	}
	if testErr := test.AssertFields(t, fieldTests, prefix, ""); testErr != nil {
		t.Fatal(testErr)
	}
}

func TestHub_Run__register(t *testing.T) {
	// GIVEN: a Hub.
	hub := NewHub()
	go hub.Run()
	time.Sleep(time.Second)

	// AND: two clients (not connected).
	client := testClient()
	otherClient := testClient()

	// WHEN: a new client connects.
	hub.register <- &client
	hub.register <- &otherClient

	// THEN: that client is registered to the Hub.
	time.Sleep(time.Second)
	if !hub.clients[&client] {
		t.Errorf("%s\nclient wasn't registered to the Hub", packageName)
	}
}

func TestHub_Run__unregister(t *testing.T) {
	// GIVEN: a Hub.
	hub := NewHub()
	go hub.Run()
	time.Sleep(time.Second)

	// AND: a client.
	client := testClient()
	otherClient := testClient()
	hub.register <- &client
	hub.register <- &otherClient
	if !hub.clients[&client] {
		t.Errorf("%s\nclient wasn't registered to the Hub", packageName)
	}

	// WHEN: that client disconnects.
	hub.unregister <- &client
	hub.unregister <- &otherClient

	// THEN: that client is unregistered to the Hub.
	if hub.clients[&client] {
		t.Errorf(
			"%s\nHub.Run() client should have been removed from the Hub after unregister\nclients: %v",
			packageName, hub.clients,
		)
	}
}

func TestHub_Run__broadcast(t *testing.T) {
	// GIVEN: a Hub.
	client := testClient()
	hub := client.hub
	go hub.Run()

	// AND: a client.
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(2 * time.Second)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}

	// WHEN: that message is broadcast.
	data, _ := decode.Marshal("json", msg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN: that message is broadcast to the client.
	got := <-client.send
	var gotMsg AnnounceMSG
	_ = decode.Unmarshal("json", got, &gotMsg)
	if gotMsg != msg {
		t.Errorf(
			"%s\nHub.Run() message sent to Broadcast channel should have been received by the client channel\ngot:  %v\nwant: %v",
			packageName, gotMsg, msg,
		)
	}
}

func TestHub_Run__broadcast_allClients(t *testing.T) {
	// GIVEN: a Hub.
	hub := NewHub()
	go hub.Run()
	time.Sleep(time.Second)

	// AND: multiple clients.
	clientA := testClient()
	clientB := testClient()
	clientC := testClient()
	hub.register <- &clientA
	hub.register <- &clientB
	hub.register <- &clientC
	time.Sleep(2 * time.Second)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf("%s\nHub.Run() failed to marshal broadcast message: %v", packageName, err)
	}

	// WHEN: that message is broadcast.
	hub.Broadcast <- data
	time.Sleep(time.Second)

	prefix := fmt.Sprintf("%s\nHub.Run()", packageName)

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

func TestHub_Run__broadcast_dropsFullClient(t *testing.T) {
	// GIVEN: a hub.
	hub := NewHub()
	go hub.Run()
	time.Sleep(time.Second)

	// AND: a client with a full outbound buffer and another with capacity.
	slowClient := Client{
		ip:   "1.1.1.1",
		send: make(chan []byte, 1),
	}
	slowClient.send <- []byte(`{"type":"test"}`)
	readyClient := testClient()
	hub.register <- &slowClient
	hub.register <- &readyClient
	time.Sleep(2 * time.Second)

	// AND: a valid message.
	msg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, err := decode.Marshal("json", msg)
	if err != nil {
		t.Fatalf("%s\nHub.Run() failed to marshal broadcast message: %v", packageName, err)
	}

	// WHEN: that message is broadcast.
	hub.Broadcast <- data
	time.Sleep(time.Second)

	prefix := fmt.Sprintf("%s\nHub.Run()", packageName)

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

func TestHub_Run__broadcast_invalid(t *testing.T) {
	// GIVEN: a Hub.
	client := testClient()
	hub := client.hub
	go hub.Run()
	time.Sleep(time.Second)

	// AND: a Client.
	hub.register <- &client
	time.Sleep(2 * time.Second)

	// AND: an invalid message.
	msg := []byte("key: value\nkey: value")

	// WHEN: that message is broadcast.
	data, _ := decode.Marshal("json", msg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN: that message is NOT sent to the client.
	got := len(client.send)
	want := 0
	if got != want {
		t.Errorf(
			"%s\nHub.Run() message sent to the Broadcast channel should have failed Unmarshal and not been sent\n"+
				"got:  %d\nwant: %d",
			packageName, got, want,
		)
	}
}
