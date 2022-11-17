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

//go:build unit

package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/release-argus/Argus/util"
)

func TestNewHub(t *testing.T) {
	// GIVEN we want a WebSocket Hub

	// WHEN we create a new one with NewHub
	hub := NewHub()

	// THEN it returns a Hub with all the channels and maps initialised
	if hub.Broadcast == nil {
		t.Error("hub.Broadcast should be non-nil")
	}
	if hub.register == nil {
		t.Error("hub.register should be non-nil")
	}
	if hub.unregister == nil {
		t.Error("hub.unregister should be non-nil")
	}
	if hub.clients == nil {
		t.Error("hub.clients should be non-nil")
	}
}

func TestRunWithRegister(t *testing.T) {
	// GIVEN a WebSocket Hub and API
	hub := NewHub()
	api := API{}
	go hub.Run(util.NewJLog("WARN", false))
	time.Sleep(time.Second)

	// WHEN a new client connects
	client := testClient()
	client.api = &api
	client.hub = hub
	hub.register <- &client
	time.Sleep(time.Second)

	// THEN that client is registered to the Hub
	if !hub.clients[&client] {
		t.Error("Client wasn't registerd to the Hub")
	}
}

func TestRunWithUnregister(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub
	client := testClient()
	hub := client.hub
	go hub.Run(util.NewJLog("WARN", false))
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(time.Second)

	// WHEN that client disconnects
	hub.unregister <- &client
	hub.unregister <- &client

	// THEN that client is unregistered to the Hub
	if hub.clients[&client] {
		t.Errorf("Client should have been removed from the Hub\n%v",
			hub.clients)
	}
}

func TestRunWithBroadcast(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub
	// and a valid message wants to be sent
	client := testClient()
	hub := client.hub
	go hub.Run(util.NewJLog("WARN", false))
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(2 * time.Second)

	// WHEN that message is Broadcast
	sentMsg := AnnounceMSG{
		Type:      "test",
		ServiceID: "something",
	}
	data, _ := json.Marshal(sentMsg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN that message is broadcast to the client
	got := <-client.send
	var gotMsg AnnounceMSG
	json.Unmarshal(got, &gotMsg)
	if gotMsg != sentMsg {
		t.Errorf("Client should have received %v, not %v",
			sentMsg, gotMsg)
	}
}

func TestRunWithInvalidBroadcast(t *testing.T) {
	// GIVEN a Client is connected to the WebSocket Hub
	// and an invalid message wants to be sent
	client := testClient()
	hub := client.hub
	go hub.Run(util.NewJLog("WARN", false))
	time.Sleep(time.Second)
	hub.register <- &client
	time.Sleep(time.Second)

	// WHEN that message is Broadcast
	sentMsg := []byte("key: value\nkey: value")
	data, _ := json.Marshal(sentMsg)
	hub.Broadcast <- data
	time.Sleep(time.Second)

	// THEN that message is broadcast to the client
	got := len(client.send)
	want := 0
	if got != want {
		t.Errorf("Sent message should have failed Unmarshal and notohing sent. Got %d messages",
			got)
	}
}
