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

package conversions

import (
	"fmt"
	"testing"
)

func TestConversionGotifyNil(t *testing.T) {
	// GIVEN a nil Gotify
	var gotify *Gotify

	// WHEN Convert is called on it
	converted := gotify.Convert("")

	// THEN nothing changes
	if converted.ID != nil {
		t.Fatalf("Convert of nil changed nil to %v", converted)
	}
}

func TestConversionGotify(t *testing.T) {
	var (
		url      string = "http://mock_host:123/test"
		host     string = "mock_host"
		port     string = "123"
		path     string = "test"
		token    string = "super_secret"
		title    string = "Fancy title"
		message  string = "foo"
		priority int    = 3
		delay    string = "1s"
		maxTries uint   = 5
	)
	test := Gotify{
		URL:      &url,
		Token:    &token,
		Title:    &title,
		Message:  &message,
		Priority: &priority,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted := test.Convert("")

	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedGotify.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedGotify.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedGotify.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
	if !(token == converted.GetSelfURLField("token")) {
		t.Fatalf(`convertedGotify.URLFields.token = %q, want match for %q`, converted.GetSelfURLField("token"), token)
	}
	if !(title == converted.GetSelfParam("title")) {
		t.Fatalf(`convertedGotify.Params.title = %q, want match for %q`, converted.GetSelfURLField("title"), title)
	}
	if !(message == converted.GetSelfOption("message")) {
		t.Fatalf(`convertedGotify.Options.message = %q, want match for %q`, converted.GetSelfURLField("message"), message)
	}
	if !(fmt.Sprint(priority) == converted.GetSelfParam("priority")) {
		t.Fatalf(`convertedGotify.Params.priority = %q, want match for %q`, converted.GetSelfURLField("priority"), priority)
	}
	if !(delay == converted.GetSelfOption("delay")) {
		t.Fatalf(`convertedGotify.Options.delay = %q, want match for %q`, converted.GetSelfURLField("delay"), delay)
	}
	if !(fmt.Sprint(maxTries) == converted.GetSelfOption("max_tries")) {
		t.Fatalf(`convertedGotify.Options.max_tries = %q, want match for %q`, converted.GetSelfURLField("max_tries"), maxTries)
	}

	url = "mock_host:123"
	test = Gotify{
		URL:      &url,
		Token:    &token,
		Title:    &title,
		Message:  &message,
		Priority: &priority,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted = test.Convert("")
	path = ""
	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedGotify.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedGotify.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedGotify.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}

	url = "https://mock_host"
	test = Gotify{
		URL:      &url,
		Token:    &token,
		Title:    &title,
		Message:  &message,
		Priority: &priority,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted = test.Convert("")
	port = "443"
	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedGotify.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedGotify.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedGotify.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}

	url = "mock_host"
	test = Gotify{
		URL:      &url,
		Token:    &token,
		Title:    &title,
		Message:  &message,
		Priority: &priority,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted = test.Convert("")
	port = "80"
	if !(url == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedGotify.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), url)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedGotify.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedGotify.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
}
