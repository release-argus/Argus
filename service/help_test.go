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
	"fmt"

	db_types "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/service/latest_version"
	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/service/options"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

func stringPtr(val string) *string {
	return &val
}
func boolPtr(val bool) *bool {
	return &val
}
func stringifyPointer[T comparable](ptr *T) string {
	str := "nil"
	if ptr != nil {
		str = fmt.Sprint(*ptr)
	}
	return str
}
func testLogging() {
	jLog = utils.NewJLog("WARN", false)
	jLog.Testing = true
	var webhookLogs *webhook.Slice
	webhookLogs.Init(jLog, nil, nil, nil, nil, nil, nil, nil)
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
				RegexContent: "content",
				RegexVersion: "version",
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		Dashboard: DashboardOptions{
			AutoApprove: boolPtr(false),
			Icon:        "test",
			IconLinkTo:  "https://example.com",
			WebURL:      "https://release-argus.io",
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
		announceChannel chan []byte           = make(chan []byte, 5)
		saveChannel     chan bool             = make(chan bool, 5)
		databaseChannel chan db_types.Message = make(chan db_types.Message, 5)
	)
	svc := Service{
		ID:   "test",
		Type: "url",
		LatestVersion: latest_version.Lookup{
			URL: "release-argus/Argus",
			Require: &filters.Require{
				RegexContent: "content",
				RegexVersion: "version",
			},
			AllowInvalidCerts: boolPtr(true),
			UsePreRelease:     boolPtr(false),
		},
		Dashboard: DashboardOptions{
			AutoApprove:  boolPtr(false),
			Icon:         "test",
			IconLinkTo:   "https://release-argus.io",
			WebURL:       "https://release-argus.io",
			Defaults:     &DashboardOptions{},
			HardDefaults: &DashboardOptions{},
		},
		Status: service_status.Status{
			ServiceID:                stringPtr("test"),
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

func testWebHookSuccessful() *webhook.WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	return &webhook.WebHook{
		ID:                "test",
		Type:              stringPtr("github"),
		URL:               stringPtr("https://valid.release-argus.io/hooks/github-style"),
		Secret:            stringPtr("argus"),
		AllowInvalidCerts: boolPtr(false),
		DesiredStatusCode: &desiredStatusCode,
		Delay:             stringPtr("0s"),
		SilentFails:       boolPtr(false),
		MaxTries:          &whMaxTries,
		ParentInterval:    stringPtr("12m"),
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}

func testWebHookFailing() *webhook.WebHook {
	desiredStatusCode := 0
	whMaxTries := uint(1)
	return &webhook.WebHook{
		ID:                "test",
		Type:              stringPtr("github"),
		URL:               stringPtr("https://valid.release-argus.io/hooks/github-style"),
		Secret:            stringPtr("notArgus"),
		AllowInvalidCerts: boolPtr(false),
		DesiredStatusCode: &desiredStatusCode,
		Delay:             stringPtr("0s"),
		SilentFails:       boolPtr(false),
		MaxTries:          &whMaxTries,
		ParentInterval:    stringPtr("12m"),
		Main:              &webhook.WebHook{},
		Defaults:          &webhook.WebHook{},
		HardDefaults:      &webhook.WebHook{},
	}
}
