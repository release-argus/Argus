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

//go:build testing

package v1

import (
	"fmt"
	"testing"

	"github.com/gorilla/websocket"
	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/config"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/service"
	"github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func boolPtr(val bool) *bool {
	return &val
}
func stringPtr(val string) *string {
	return &val
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}

func testClient() Client {
	hub := NewHub()
	api := API{}
	return Client{
		api:  &api,
		hub:  hub,
		ip:   "1.1.1.1",
		conn: &websocket.Conn{},
		send: make(chan []byte, 5),
	}
}

func testAPI() API {
	var serviceID string = "test"
	return API{
		Config: &config.Config{
			Service: service.Slice{
				serviceID: &service.Service{
					ID:      serviceID,
					Comment: "foo",
					LatestVersion: latest_version.Lookup{
						Type: "github",
						URL:  "release-argus/Argus",
					},
				},
			},
		},
		Log: utils.NewJLog("WARN", false),
	}
}

func testService(id string) service.Service {
	return service.Service{
		ID:      id,
		Comment: "foo",
		LatestVersion: latest_version.Lookup{
			Type: "github",
			URL:  "release-argus/Argus",
		},
	}
}

func TestConvertDeployedVersionLookupToApiTypeDeployedVersionLookupWithNil(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *deployed_version.Lookup

	// WHEN convertDeployedVersionLookupToApiTypeDeployedVersionLookup is called on it
	got := convertDeployedVersionLookupToApiTypeDeployedVersionLookup(dvl)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			dvl, got)
	}
}

func TestConvertDeployedVersionLookupToApiTypeDeployedVersionLookupDidConcealNasicAuthPassword(t *testing.T) {
	// GIVEN a DeployedVersionLookup with a basic auth password
	basicAuth := deployed_version.BasicAuth{
		Username: "username",
		Password: "pass123",
	}
	dvl := deployed_version.Lookup{
		URL:       "https://example.com",
		BasicAuth: &basicAuth,
	}

	// WHEN convertDeployedVersionLookupToApiTypeDeployedVersionLookup is called on it
	got := convertDeployedVersionLookupToApiTypeDeployedVersionLookup(&dvl)

	// THEN basic auth was censored
	want := api_types.BasicAuth{
		Username: dvl.BasicAuth.Username,
		Password: "<secret>",
	}
	if got.BasicAuth.Username != want.Username ||
		got.BasicAuth.Password != want.Password {
		t.Errorf("BasicAuth was not carried over with a concealed password\nfrom: %v\nto:   %v",
			dvl, got)
	}
}

func TestConvertDeployedVersionLookupToApiTypeDeployedVersionLookupDidConcealHeaderKeys(t *testing.T) {
	// GIVEN a DeployedVersionLookup with headers
	dvl := deployed_version.Lookup{
		URL: "https://example.com",
		Headers: []deployed_version.Header{
			{Key: "X-Test-0", Value: "foo"},
			{Key: "X-Test-1", Value: "foo"},
		},
	}

	// WHEN convertDeployedVersionLookupToApiTypeDeployedVersionLookup is called on it
	got := convertDeployedVersionLookupToApiTypeDeployedVersionLookup(&dvl)

	// THEN the header keys were censored
	want := []api_types.Header{
		{Key: dvl.Headers[0].Key, Value: "<secret>"},
		{Key: dvl.Headers[1].Key, Value: "<secret>"},
	}
	if len(got.Headers) != 2 ||
		got.Headers[0] != want[0] ||
		got.Headers[1] != want[1] {
		t.Errorf("header keys should have been censored\nfrom: %v\nto:   %v",
			dvl, *got)
	}
}

func TestConvertDeployedVersionLookupToApiTypeDeployedVersionLookup(t *testing.T) {
	// GIVEN a DeployedVersionLookup with a basic auth password
	allowInvalidCerts := true
	dvl := deployed_version.Lookup{
		URL:               "https://example.com",
		AllowInvalidCerts: &allowInvalidCerts,
		JSON:              "foo",
		Regex:             "bar",
	}

	// WHEN convertDeployedVersionLookupToApiTypeDeployedVersionLookup is called on it
	got := convertDeployedVersionLookupToApiTypeDeployedVersionLookup(&dvl)

	// THEN the vars were placed correctly
	if got.URL != dvl.URL ||
		*got.AllowInvalidCerts != *dvl.AllowInvalidCerts ||
		got.JSON != dvl.JSON ||
		got.Regex != dvl.Regex {
		t.Errorf("conversion of %v should have given nil, not %v",
			dvl, got)
	}
}

func TestConvertURLCommandSliceToAPITypeURLCommandSliceWithNil(t *testing.T) {
	// GIVEN a nil URL Command slice
	var slice *filters.URLCommandSlice

	// WHEN convertURLCommandSliceToAPITypeURLCommandSlice is called on it
	got := convertURLCommandSliceToAPITypeURLCommandSlice(slice)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			slice, got)
	}
}

func TestConvertURLCommandSliceToAPITypeURLCommandSlice(t *testing.T) {
	// GIVEN a URL Command slice
	slice := filters.URLCommandSlice{
		{
			Type:  "something",
			Regex: stringPtr("foo"),
			Index: 0,
			Text:  stringPtr("bish"),
			Old:   stringPtr("bash"),
			New:   stringPtr("bosh"),
		},
		{
			Type:  "another",
			Regex: stringPtr("bar"),
			Index: 1,
			Text:  stringPtr("bosh"),
			Old:   stringPtr("bish"),
			New:   stringPtr("bash"),
		},
	}

	// WHEN convertURLCommandSliceToAPITypeURLCommandSlice is called on it
	got := convertURLCommandSliceToAPITypeURLCommandSlice(&slice)

	// THEN the slice was converted correctly
	if len(got) != len(slice) ||
		slice[0].Regex != got[0].Regex ||
		slice[0].Index != got[0].Index ||
		slice[1].Regex != got[1].Regex ||
		slice[1].Index != got[1].Index {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	}
}

func TestConvertNotifySliceToAPITypeNotifySliceWithNil(t *testing.T) {
	// GIVEN a nil Notify slice
	var slice *shoutrrr.Slice

	// WHEN convertNotifySliceToAPITypeNotifySlice is called on it
	got := convertNotifySliceToAPITypeNotifySlice(slice)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			slice, got)
	}
}

func TestConvertNotifySliceToAPITypeNotifySlice(t *testing.T) {
	// GIVEN a Notify slice
	slice := shoutrrr.Slice{
		"test": {
			Type: "discord",
			Options: map[string]string{
				"message": "something {{ version }}",
			},
			URLFields: map[string]string{
				"port":  "25",
				"other": "something",
			},
			Params: map[string]string{
				"avatar": "fizz",
			},
		},
		"other": {
			Type: "slack",
			Options: map[string]string{
				"message": "foo {{ version }}",
				"delay":   "something",
			},
			URLFields: map[string]string{
				"port": "8080",
			},
			Params: map[string]string{
				"avatar": "buz",
				"other":  "something",
			},
		},
	}

	// WHEN convertNotifySliceToAPITypeNotifySlice is called on it
	got := convertNotifySliceToAPITypeNotifySlice(&slice)

	// THEN the slice was converted correctly
	if len(*got) != len(slice) {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	} else {
		for i := range *got {
			for j := range (*got)[i].Options {
				if (*got)[i].Options[j] != slice[i].Options[j] ||
					len((*got)[i].Options[j]) != len(slice[i].Options[j]) {
					t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
						slice, got)
				}
			}
			for j := range (*got)[i].URLFields {
				if (*got)[i].URLFields[j] != slice[i].URLFields[j] ||
					len((*got)[i].URLFields[j]) != len(slice[i].URLFields[j]) {
					t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
						slice, got)
				}
			}
			for j := range (*got)[i].Params {
				if (*got)[i].Params[j] != slice[i].Params[j] ||
					len((*got)[i].Params[j]) != len(slice[i].Params[j]) {
					t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
						slice, got)
				}
			}
		}
	}
}

func TestConvertNotifySliceToAPITypeNotifySliceDoesCensor(t *testing.T) {
	// GIVEN a Notify slice
	slice := shoutrrr.Slice{
		"test": {
			Type: "discord",
			Options: map[string]string{
				"message": "something {{ version }}",
			},
			URLFields: map[string]string{
				"port":   "25",
				"apikey": "fizz",
			},
			Params: map[string]string{
				"avatar":  "argus",
				"devices": "buzz",
			},
		},
	}

	// WHEN convertNotifySliceToAPITypeNotifySlice is called on it
	got := convertNotifySliceToAPITypeNotifySlice(&slice)

	// THEN the slice was converted correctly
	if (*got)["test"].URLFields["port"] != slice["test"].URLFields["port"] ||
		(*got)["test"].URLFields["apikey"] != "<secret>" ||
		(*got)["test"].Params["avatar"] != slice["test"].Params["avatar"] ||
		(*got)["test"].Params["devices"] != "<secret>" {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	}
}

func TestConvertCommandSliceToAPITypeCommandSliceWithNil(t *testing.T) {
	// GIVEN a nil Command slice
	var slice *command.Slice

	// WHEN convertCommandSliceToAPITypeCommandSlice is called on it
	got := convertCommandSliceToAPITypeCommandSlice(slice)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			slice, got)
	}
}

func TestConvertCommandSliceToAPITypeCommandSlice(t *testing.T) {
	// GIVEN a Command slice
	slice := command.Slice{
		{"ls", "-lah"},
		{"/bin/bash", "something.sh"},
	}

	// WHEN convertCommandSliceToAPITypeCommandSlice is called on it
	got := convertCommandSliceToAPITypeCommandSlice(&slice)

	// THEN the slice was converted correctly
	if len(*got) != len(slice) {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	} else {
		for i := range *got {
			for j := range (*got)[i] {
				if (*got)[i][j] != slice[i][j] {
					t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
						slice, got)
				}
			}
		}
	}
}

func TestConvertWebHookToAPITypeWebHookWithNil(t *testing.T) {
	// GIVEN a nil WebHook
	var slice *webhook.WebHook

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(slice)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			slice, got)
	}
}

func TestConvertWebHookToAPITypeWebHook(t *testing.T) {
	// GIVEN a WebHook
	wh := webhook.WebHook{
		URL: "https://example.com",
	}

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(&wh)

	// THEN the slice was converted correctly
	if wh.URL != *got.URL {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			wh, got)
	}
}

func TestConvertWebHookToAPITypeWebHookDidCensorSecret(t *testing.T) {
	// GIVEN a WebHook
	wh := webhook.WebHook{
		URL:    "https://example.com",
		Secret: "shazam",
	}

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(&wh)

	// THEN the slice was converted correctly
	want := "<secret>"
	if *got.Secret != want {
		t.Errorf("secret was not censored to the expected %q. Got %q",
			want, *got.Secret)
	}
}

func TestConvertWebHookToAPITypeWebHookDidCopyHeaders(t *testing.T) {
	// GIVEN a WebHook
	var (
		headers = map[string]string{
			"X-Something": "foo",
		}
	)
	wh := webhook.WebHook{
		URL:           "https://example.com",
		CustomHeaders: headers,
	}

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(&wh)

	// THEN the slice was converted correctly
	if (*got.CustomHeaders)["X-Something"] != wh.CustomHeaders["X-Something"] {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			wh, got)
	}
}

func TestConvertWebHookSliceToAPITypeWebHookSliceWithNil(t *testing.T) {
	// GIVEN a nil WebHook slice
	var slice *webhook.Slice

	// WHEN convertWebHookSliceToAPITypeWebHookSlice is called on it
	got := convertWebHookSliceToAPITypeWebHookSlice(slice)

	// THEN nil was returned
	if got != nil {
		t.Errorf("conversion of %v should have given nil, not %v",
			slice, got)
	}
}

func TestConvertWebHookSliceToAPITypeWebHookSlice(t *testing.T) {
	// GIVEN a WebHook slice
	slice := webhook.Slice{
		"test":  {URL: "https://example.com"},
		"other": {URL: "https://release-argus.io"},
	}

	// WHEN convertWebHookSliceToAPITypeWebHookSlice is called on it
	got := convertWebHookSliceToAPITypeWebHookSlice(&slice)

	// THEN the slice was converted correctly
	if len(*got) != len(slice) {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	} else {
		for i := range *got {
			if stringifyPointer((*got)[i].URL) != slice[i].URL {
				t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
					slice, got)
			}
		}
	}
}
