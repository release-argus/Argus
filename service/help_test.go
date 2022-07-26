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
	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/webhook"
)

func stringPtr(val string) *string {
	return &val
}
func boolPtr(val bool) *bool {
	return &val
}

func testServiceGitHub() Service {
	var (
		announceChannel chan []byte           = make(chan []byte, 2)
		saveChannel     chan bool             = make(chan bool, 5)
		databaseChannel chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := Service{
		ID:   "test",
		Type: "github",
		LatestVersion: latest_version.Lookup{
			AccessToken: stringPtr("secret"),
			URL:         "release-argus/Argus",
			Require: &filters.Require{
				RegexContent: stringPtr("content"),
				RegexVersion: stringPtr("version"),
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		Dashboard: DashboardOptions{
			AutoApprove: boolPtr(false),
			Icon:        stringPtr("test"),
			IconLinkTo:  stringPtr("https://example.com"),
			WebURL:      stringPtr("https://release-argus.io"),
		},
		Status: service_status.Status{
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
			AnnounceChannel:          &announceChannel,
			DatabaseChannel:          &databaseChannel,
			SaveChannel:              &saveChannel,
		},
		Options: options.Options{
			Interval:           stringPtr("5s"),
			SemanticVersioning: boolPtr(true),
		},
		Defaults: &Service{},
		HardDefaults: &Service{
			Active: boolPtr(true),
		},
	}
	svc.Status.WebURL = &svc.Dashboard.WebURL
	return svc
}

func testServiceURL() Service {
	var (
		announceChannel chan []byte           = make(chan []byte, 2)
		saveChannel     chan bool             = make(chan bool, 5)
		databaseChannel chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := Service{
		ID:   "test",
		Type: "url",
		LatestVersion: latest_version.Lookup{
			URL: "release-argus/Argus",
			Require: &filters.Require{
				RegexContent: stringPtr("content"),
				RegexVersion: stringPtr("version"),
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		Dashboard: DashboardOptions{
			AutoApprove: boolPtr(false),
			Icon:        stringPtr("test"),
			IconLinkTo:  stringPtr("https://release-argus.io"),
			WebURL:      stringPtr("https://release-argus.io"),
		},
		Status: service_status.Status{
			ApprovedVersion:          "1.1.1",
			LatestVersion:            "2.2.2",
			LatestVersionTimestamp:   "2002-02-02T02:02:02Z",
			DeployedVersion:          "0.0.0",
			DeployedVersionTimestamp: "2001-01-01T01:01:01Z",
			AnnounceChannel:          &announceChannel,
			DatabaseChannel:          &databaseChannel,
			SaveChannel:              &saveChannel,
		},
		Options: options.Options{
			Interval:           stringPtr("5s"),
			SemanticVersioning: boolPtr(true),
		},
		Defaults: &Service{},
		HardDefaults: &Service{
			Active: boolPtr(true),
		},
	}
	svc.Status.WebURL = &svc.Dashboard.WebURL
	return svc
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
	parentInterval := "12m"
	return webhook.WebHook{
		ID:                whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		ParentInterval:    &parentInterval,
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
	parentInterval := "12m"
	return webhook.WebHook{
		ID:                whID,
		Type:              &whType,
		URL:               &whURL,
		Secret:            &whSecret,
		AllowInvalidCerts: &whAllowInvalidCerts,
		DesiredStatusCode: &whDesiredStatusCode,
		SilentFails:       &whSilentFails,
		MaxTries:          &whMaxTries,
		ParentInterval:    &parentInterval,
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}
