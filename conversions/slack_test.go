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

func TestConversionSlack(t *testing.T) {
	var (
		url       string = "mock_host:123/test/hooks/fez"
		host      string = "mock_host"
		port      string = "123"
		path      string = "test"
		token     string = "fez"
		iconEmoji string = "piranha"
		iconURL   string = "example.com/icon.png"
		username  string = "bar"
		message   string = "foo"
		delay     string = "1s"
		maxTries  uint   = 5
		channel   string = "webhook"
		cType     string = "mattermost"
	)
	test := Slack{
		URL:       &url,
		Username:  &username,
		Message:   &message,
		Delay:     &delay,
		MaxTries:  &maxTries,
		IconEmoji: &iconEmoji,
	}
	converted := test.Convert("", url)

	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedSlack.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedSlack.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedSlack.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
	if !(token == converted.GetSelfURLField("token")) {
		t.Fatalf(`convertedSlack.URLFields.token = %q, want match for %q`, converted.GetSelfURLField("token"), token)
	}
	if !(username == converted.GetSelfParam("botname")) {
		t.Fatalf(`convertedSlack.Params.botname = %q, want match for %q`, converted.GetSelfURLField("botname"), username)
	}
	if !(iconEmoji == converted.GetSelfParam("icon")) {
		t.Fatalf(`convertedSlack.Params.icon = %q, want match for %q`, converted.GetSelfURLField("icon"), iconEmoji)
	}
	if !(message == converted.GetSelfOption("message")) {
		t.Fatalf(`convertedSlack.Options.message = %q, want match for %q`, converted.GetSelfURLField("message"), message)
	}
	if !(delay == converted.GetSelfOption("delay")) {
		t.Fatalf(`convertedSlack.Options.delay = %q, want match for %q`, converted.GetSelfURLField("delay"), delay)
	}
	if !(fmt.Sprint(maxTries) == converted.GetSelfOption("max_tries")) {
		t.Fatalf(`convertedSlack.Options.max_tries = %q, want match for %d`, converted.GetSelfURLField("max_tries"), maxTries)
	}
	if !(cType == converted.Type) {
		t.Fatalf(`convertedSlack.Type = %q, want match for %q`, converted.Type, cType)
	}

	url = "mock_host:123/hooks/fez"
	test = Slack{
		URL:      &url,
		Username: &username,
		Message:  &message,
		Delay:    &delay,
		MaxTries: &maxTries,
		IconURL:  &iconURL,
	}
	converted = test.Convert("", url)
	path = ""
	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedSlack.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedSlack.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedSlack.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
	if !(token == converted.GetSelfURLField("token")) {
		t.Fatalf(`convertedSlack.URLFields.token = %q, want match for %q`, converted.GetSelfURLField("token"), token)
	}
	if !(iconURL == converted.GetSelfParam("icon")) {
		t.Fatalf(`convertedSlack.Params.icon = %q, want match for %q`, converted.GetSelfParam("icon"), iconURL)
	}

	url = "mock_host/hooks/fez"
	test = Slack{
		URL:      &url,
		Username: &username,
		Message:  &message,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted = test.Convert("", url)
	port = "80"
	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedSlack.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedSlack.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedSlack.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
	if !(token == converted.GetSelfURLField("token")) {
		t.Fatalf(`convertedSlack.URLFields.token = %q, want match for %q`, converted.GetSelfURLField("token"), token)
	}

	url = "https://hooks.slack.com/xxxx/yyyy/zzzz"
	test = Slack{
		URL:      &url,
		Username: &username,
		Message:  &message,
		Delay:    &delay,
		MaxTries: &maxTries,
	}
	converted = test.Convert("", url)
	host = ""
	port = ""
	token = "hook:xxxx-yyyy-zzzz"
	if !(host == converted.GetSelfURLField("host")) {
		t.Fatalf(`convertedSlack.URLFields.host = %q, want match for %q`, converted.GetSelfURLField("host"), host)
	}
	if !(port == converted.GetSelfURLField("port")) {
		t.Fatalf(`convertedSlack.URLFields.port = %q, want match for %q`, converted.GetSelfURLField("port"), port)
	}
	if !(path == converted.GetSelfURLField("path")) {
		t.Fatalf(`convertedSlack.URLFields.path = %q, want match for %q`, converted.GetSelfURLField("path"), path)
	}
	if !(token == converted.GetSelfURLField("token")) {
		t.Fatalf(`convertedSlack.URLFields.token = %q, want match for %q`, converted.GetSelfURLField("token"), token)
	}
	if !(channel == converted.GetSelfURLField("channel")) {
		t.Fatalf(`convertedSlack.URLFields.channel = %q, want match for %q`, converted.GetSelfURLField("channel"), channel)
	}
}
