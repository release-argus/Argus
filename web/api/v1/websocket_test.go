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

package v1

import (
	"testing"

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/notifiers/shoutrrr"
	deployed_version_lookup "github.com/release-argus/Argus/service/deployed_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	api_types "github.com/release-argus/Argus/web/api/types"
	"github.com/release-argus/Argus/webhook"
)

func TestConvertDeployedVersionLookupToApiTypeDeployedVersionLookupWithNil(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *deployed_version_lookup.Lookup

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
	basicAuth := deployed_version_lookup.BasicAuth{
		Username: "username",
		Password: "pass123",
	}
	dvl := deployed_version_lookup.Lookup{
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
	dvl := deployed_version_lookup.Lookup{
		URL: "https://example.com",
		Headers: []deployed_version_lookup.Header{
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
	dvl := deployed_version_lookup.Lookup{
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
	uType0 := "0something"
	regex0 := "0foo"
	text0 := "0bish"
	old0 := "0bash"
	uNew0 := "0bosh"
	uType1 := "1something"
	regex1 := "1foo"
	text1 := "1bish"
	old1 := "1bash"
	uNew1 := "1bosh"
	ignoreMisses := true
	slice := filters.URLCommandSlice{
		{
			Type:         uType0,
			Regex:        &regex0,
			Index:        0,
			Text:         &text0,
			Old:          &old0,
			New:          &uNew0,
			IgnoreMisses: &ignoreMisses,
		},
		{
			Type:         uType1,
			Regex:        &regex1,
			Index:        1,
			Text:         &text1,
			Old:          &old1,
			New:          &uNew1,
			IgnoreMisses: &ignoreMisses,
		},
	}

	// WHEN convertURLCommandSliceToAPITypeURLCommandSlice is called on it
	got := convertURLCommandSliceToAPITypeURLCommandSlice(&slice)

	// THEN the slice was converted correctly
	if len(*got) != len(slice) ||
		slice[0].Regex != (*got)[0].Regex ||
		slice[0].Index != (*got)[0].Index ||
		slice[0].IgnoreMisses != (*got)[0].IgnoreMisses ||
		slice[1].Regex != (*got)[1].Regex ||
		slice[1].Index != (*got)[1].Index ||
		slice[1].IgnoreMisses != (*got)[1].IgnoreMisses {
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
	var (
		url = "https://example.com"
	)
	wh := webhook.WebHook{
		URL: &url,
	}

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(&wh)

	// THEN the slice was converted correctly
	if wh.URL != got.URL {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			wh, got)
	}
}

func TestConvertWebHookToAPITypeWebHookDidCensorSecret(t *testing.T) {
	// GIVEN a WebHook
	var (
		url    = "https://example.com"
		secret = "shazam"
	)
	wh := webhook.WebHook{
		URL:    &url,
		Secret: &secret,
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
		url     = "https://example.com"
		headers = map[string]string{
			"X-Something": "foo",
		}
	)
	wh := webhook.WebHook{
		URL:           &url,
		CustomHeaders: &headers,
	}

	// WHEN convertWebHookToAPITypeWebHook is called on it
	got := convertWebHookToAPITypeWebHook(&wh)

	// THEN the slice was converted correctly
	if (*got.CustomHeaders)["X-Something"] != (*wh.CustomHeaders)["X-Something"] {
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
	var (
		url0 = "https://example.com"
		url1 = "https://release-argus.io"
	)
	slice := webhook.Slice{
		"test":  {URL: &url0},
		"other": {URL: &url1},
	}

	// WHEN convertWebHookSliceToAPITypeWebHookSlice is called on it
	got := convertWebHookSliceToAPITypeWebHookSlice(&slice)

	// THEN the slice was converted correctly
	if len(*got) != len(slice) {
		t.Errorf("converted incorrectly\nfrom: %v\nto:   %v",
			slice, got)
	} else {
		for i := range *got {
			if (*got)[i].URL != slice[i].URL {
				t.Fatalf("converted incorrectly\nfrom: %v\nto:   %v",
					slice, got)
			}
		}
	}
}
