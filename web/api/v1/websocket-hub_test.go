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

//go:build unit

package v1

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	// GIVEN we want a WebSocket Hub.

	// WHEN we create a new one with NewHub.
	hub := NewHub()

	// THEN it returns a Hub with all the channels and maps initialised.
	if hub.Broadcast == nil {
		t.Errorf("%s\nhub.Broadcast should be non-nil",
			packageName)
	}
	if hub.register == nil {
		t.Errorf("%s\nhub.register should be non-nil",
			packageName)
	}
	if hub.unregister == nil {
		t.Errorf("%s\nhub.unregister should be non-nil",
			packageName)
	}
	if hub.clients == nil {
		t.Errorf("%s\nhub.clients should be non-nil",
			packageName)
	}
}

func TestHub_RunWithRegister(t *testing.T) {
	// GIVEN a WebSocket Hub and two clients.
	hub := NewHub()
	go hub.Run()
	client := testClient()
	otherClient := testClient()

	// WHEN a new client connects (two for synchronisation).
	hub.register <- &client
	hub.register <- &otherClient
	hub.register <- &otherClient

	// THEN that client is registered to the Hub.
	// DATA RACE - Unsure why as register is a second before this read.
	if !hub.clients[&client] {
		t.Errorf("%s\nclient wasn't registered to the Hub",
			packageName)
	}
}

func TestHub_RunWithUnregister(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub.
	client := testClient()
	otherClient := testClient()
	hub := client.hub
	go hub.Run()
	hub.register <- &client
	hub.register <- &otherClient
	hub.register <- &otherClient
	if !hub.clients[&client] {
		t.Errorf("%s\nclient wasn't registered to the Hub",
			packageName)
	}

	// WHEN that client disconnects (two for synchronisation).
	hub.unregister <- &client
	hub.unregister <- &otherClient
	hub.unregister <- &otherClient

	// THEN that client is unregistered to the Hub.
	if hub.clients[&client] {
		t.Errorf("%s\nclient should have been removed from the Hub\nclients: %v",
			packageName, hub.clients)
	}
}

func TestHub_RunWithBroadcast(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub,
	// and a valid message wants to be sent.
	client := testClient()
	hub := client.hub
	go hub.Run()
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(2 * time.Second)

	// WHEN that message is Broadcast.
	sentMsg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, _ := json.Marshal(sentMsg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN that message is broadcast to the client.
	got := <-client.send
	var gotMsg AnnounceMSG
	json.Unmarshal(got, &gotMsg)
	if gotMsg != sentMsg {
		t.Errorf("%s\nclient message mismatch\nwant: %v\ngot:  %v",
			packageName, sentMsg, gotMsg)
	}
}

func TestHub_RunWithInvalidBroadcast(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub,
	// and an invalid message wants to be sent.
	client := testClient()
	hub := client.hub
	go hub.Run()
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(time.Second)

	// WHEN that message is Broadcast.
	sentMsg := []byte("key: value\nkey: value")
	data, _ := json.Marshal(sentMsg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN that message is broadcast to the client.
	got := len(client.send)
	want := 0
	if got != want {
		t.Errorf("%s\nsent message should have failed Unmarshal and not sent\nwant: %d messages\ngot:  %d",
			packageName, want, got)
	}
}
