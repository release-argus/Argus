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

package service

import (
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/webhook"
)

func testServiceGitHub() Service {
	var (
		sID                 string      = "test"
		sType               string      = "github"
		sAccessToken        string      = "secret"
		sURL                string      = "release-argus/Argus"
		sWebURL             string      = "https://release-argus.io"
		sRegexContent       string      = "content"
		sRegexVersion       string      = "version"
		sAnnounceChannel    chan []byte = make(chan []byte, 2)
		sAllowInvalidCerts  bool        = false
		sSemanticVersioning bool        = true
		sAutoApprove        bool        = false
		sIgnoreMisses       bool        = false
		sUsePreRelease      bool        = false
		sActive             bool        = true
		sInterval           string      = "1s"
		sSaveChannel        chan bool   = make(chan bool, 5)
	)
	return Service{
		ID:           &sID,
		Type:         &sType,
		AccessToken:  &sAccessToken,
		URL:          &sURL,
		WebURL:       &sWebURL,
		RegexContent: &sRegexContent,
		RegexVersion: &sRegexVersion,
		Status: &service_status.Status{
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
		},
		SemanticVersioning: &sSemanticVersioning,
		AllowInvalidCerts:  &sAllowInvalidCerts,
		AutoApprove:        &sAutoApprove,
		IgnoreMisses:       &sIgnoreMisses,
		Icon:               "test",
		UsePreRelease:      &sUsePreRelease,
		Announce:           &sAnnounceChannel,
		SaveChannel:        &sSaveChannel,
		Interval:           &sInterval,
		Defaults:           &Service{},
		HardDefaults: &Service{
			Active: &sActive,
		},
	}
}

func testServiceURL() Service {
	var (
		sID                 string      = "test"
		sType               string      = "url"
		sAccessToken        string      = "secret"
		sURL                string      = "https://release-argus.io"
		sWebURL             string      = "https://release-argus.io"
		sRegexContent       string      = "content"
		sRegexVersion       string      = "version"
		sAnnounceChannel    chan []byte = make(chan []byte, 2)
		sAllowInvalidCerts  bool        = false
		sSemanticVersioning bool        = true
		sAutoApprove        bool        = false
		sIgnoreMisses       bool        = false
		sUsePreRelease      bool        = false
		sActive             bool        = true
		sInterval           string      = "10s"
		sSaveChannel        chan bool   = make(chan bool, 5)
	)
	return Service{
		ID:           &sID,
		Type:         &sType,
		AccessToken:  &sAccessToken,
		URL:          &sURL,
		WebURL:       &sWebURL,
		RegexContent: &sRegexContent,
		RegexVersion: &sRegexVersion,
		Status: &service_status.Status{
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
		},
		SemanticVersioning: &sSemanticVersioning,
		AllowInvalidCerts:  &sAllowInvalidCerts,
		AutoApprove:        &sAutoApprove,
		IgnoreMisses:       &sIgnoreMisses,
		Icon:               "test",
		UsePreRelease:      &sUsePreRelease,
		Announce:           &sAnnounceChannel,
		SaveChannel:        &sSaveChannel,
		Interval:           &sInterval,
		Defaults:           &Service{},
		HardDefaults: &Service{
			Active: &sActive,
		},
	}
}

func testDeployedVersion() DeployedVersionLookup {
	var (
		allowInvalidCerts bool = false
	)
	return DeployedVersionLookup{
		URL:               "https://release-argus.io",
		AllowInvalidCerts: &allowInvalidCerts,
		Regex:             "([0-9]+) The Argus Developers",
		Defaults:          &DeployedVersionLookup{},
		HardDefaults:      &DeployedVersionLookup{},
	}
}

func testWebHookSuccessful() webhook.WebHook {
	whID := "test"
	whType := "github"
	whURL := "https://httpbin.org/anything"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whSilentFails := true
	whMaxTries := uint(1)
	return webhook.WebHook{
		ID:                &whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}

func testWebHookFailing() webhook.WebHook {
	whID := "test"
	whType := "github"
	whURL := "https://httpbin.org/hidden-basic-auth/:user/:passwd"
	whSecret := "secret"
	whAllowInvalidCerts := false
	whDesiredStatusCode := 0
	whSilentFails := true
	whMaxTries := uint(1)
	return webhook.WebHook{
		ID:                &whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}

func testURLCommandRegex() URLCommand {
	regex := "-([0-9.]+)-"
	index := 0
	ignoreMisses := false
	return URLCommand{
		Type:         "regex",
		Regex:        &regex,
		IgnoreMisses: &ignoreMisses,
		Index:        index,
	}
}

func testURLCommandReplace() URLCommand {
	old := "foo"
	new := "bar"
	return URLCommand{
		Type: "replace",
		Old:  &old,
		New:  &new,
	}
}

func testURLCommandSplit() URLCommand {
	text := "this"
	index := 1
	ignoreMisses := false
	return URLCommand{
		Type:         "split",
		Text:         &text,
		Index:        index,
		IgnoreMisses: &ignoreMisses,
	}
}
