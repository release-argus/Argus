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
	"testing"
	"time"

	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/utils"
)

func TestServiceTrackWithQueryFailRegex(t *testing.T) {
	// GIVEN a Service that fails its queries because of Regex mismatch
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.Status.LatestVersion = ""
	*service.LatestVersion.Require.RegexVersion = "beta"

	// WHEN Track is called on this Service
	want := service.Status.LatestVersion
	go service.Track()
	time.Sleep(2 * time.Second)

	// THEN the LatestVersion wont have changed
	got := service.Status.LatestVersion
	if got != want {
		t.Errorf("Query should have failed because the %q token is invalid. LatestVersion was updated from %q to %q",
			*service.AccessToken, want, got)
	}
}

func TestServiceTrackWithQueryFailSemanticVersioning(t *testing.T) {
	// GIVEN a Service that fails its queries because the version returned isn't semantic
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.Status.LatestVersion = ""

	// WHEN Track is called on this Service
	want := service.Status.LatestVersion
	go service.Track()
	time.Sleep(5 * time.Second)

	// THEN the LatestVersion wont have changed
	got := service.Status.LatestVersion
	if got != want {
		t.Errorf("Query should have failed because %q isn't semantic versioning. LatestVersion was updated from %q to %q",
			service.Status.LatestVersion, want, got)
	}
}

func TestServiceTrackWithQueryFail404(t *testing.T) {
	// GIVEN a Service that fails its queries
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.Status.DeployedVersion = "1.2.3"
	*service.URL = "https://example.undefined"

	// WHEN Track is called on this Service
	want := service.Status.LatestVersion
	go service.Track()
	time.Sleep(5 * time.Second)

	// THEN the LatestVersion wont have changed
	got := service.Status.LatestVersion
	if got != want {
		t.Errorf("Query should have failed because the %q URL is invalid. LatestVersion was updated from %q to %q",
			*service.URL, want, got)
	}
}

func TestServiceTrackWithQueryOlderFound(t *testing.T) {
	// GIVEN a Service that passes its queries
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.Status.DeployedVersion = "1.2.3"
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	service.URLCommands = &filters.URLCommandSlice{urlCommand}
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = ""

	// WHEN Track is called on this Service
	want := service.Status.LatestVersion
	go service.Track()
	time.Sleep(5 * time.Second)

	// THEN the LatestVersion wont have changed as the version found was the current latest
	got := service.Status.LatestVersion
	if got != want {
		t.Errorf("Query should have failed because %q should be older/same as %q",
			got, want)
	}
}

func TestServiceTrackWithQueryNewFound(t *testing.T) {
	// GIVEN a Service that will find a newer LatestVersion on the next Query
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.Status.LatestVersion = "0.0.0"
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	service.URLCommands = &filters.URLCommandSlice{urlCommand}
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = ""

	// WHEN Track is called on this Service
	want := "1.2.3"
	go service.Track()
	time.Sleep(5 * time.Second)

	// THEN the LatestVersion wont have changed
	got := service.Status.LatestVersion
	if got != want {
		t.Errorf("Query should have failed because the %q token is invalid. LatestVersion was updated from %q to %q",
			*service.AccessToken, want, got)
	}
}

func TestServiceSliceTrackWithInactiveServices(t *testing.T) {
	// GIVEN a Slice containing an active Service and an inactive Service
	jLog = utils.NewJLog("WARN", false)
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	service0 := testServiceURL()
	service0.Status.LatestVersion = ""
	*service0.URL = "https://release-argus.io/docs/config/service/"
	service0.URLCommands = &filters.URLCommandSlice{urlCommand}
	*service0.LatestVersion.Require.RegexContent = ""
	*service0.LatestVersion.Require.RegexVersion = ""
	active := false
	service0.Active = &active
	service1 := testServiceURL()
	service1.Status.LatestVersion = ""
	*service1.URL = "https://release-argus.io/docs/config/service/"
	service1.URLCommands = &filters.URLCommandSlice{urlCommand}
	*service1.LatestVersion.Require.RegexContent = ""
	*service1.LatestVersion.Require.RegexVersion = ""
	slice := Slice{
		"inactive": &service0,
		"active":   &service1,
	}

	// WHEN Track is called on this Service
	order := []string{"inactive", "active"}
	go slice.Track(&order)
	time.Sleep(5 * time.Second)

	// THEN the LatestVersion will only have changed for "active"
	if service1.Status.LatestVersion == "" {
		t.Errorf("service1 is active, so LatestVersion shouldn't be %q",
			service1.Status.LatestVersion)
	}
	if service0.Status.LatestVersion == service1.Status.LatestVersion {
		t.Errorf("service1's LatestVersion shouldn't be the same %q as service0's as that's inactive",
			service1.Status.LatestVersion)
	}
}
