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

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/webhook"
)

func TestSetDeployedVersion(t *testing.T) {
	// GIVEN a Service with LatestVersion == ApprovedVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion

	// WHEN SetDeployedVersion is called on it
	service.SetDeployedVersion(service.Status.LatestVersion)

	// THEN DeployedVersion is set to this version
	got := service.Status.DeployedVersion
	want := service.Status.LatestVersion
	if got != want {
		t.Errorf("Expected DeployedVersion to be set to %q, not %q",
			want, got)
	}
}

func TestSetDeployedVersionDidSetDeployedVersionTimestamp(t *testing.T) {
	// GIVEN a Service with LatestVersion == ApprovedVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion

	// WHEN SetDeployedVersion is called on it
	start := time.Now()
	service.SetDeployedVersion(service.Status.LatestVersion)

	// THEN DeployedVersionTimestamp is set to now in time
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("DeployedVersionTimestamp was %v ago, not recent enough!",
			since)
	}
}

func TestSetDeployedVersionDidResetApprovedWhenMatch(t *testing.T) {
	// GIVEN a Service with LatestVersion == ApprovedVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion

	// WHEN SetDeployedVersion is called on it with this ApprovedVersion
	service.SetDeployedVersion(service.Status.ApprovedVersion)

	// THEN ApprovedVersion is reset
	got := service.Status.ApprovedVersion
	want := ""
	if got != want {
		t.Errorf("Expected ApprovedVersion to be reset to %q, not %q",
			want, got)
	}
}

func TestSetDeployedVersionDidntResetApprovedWhenMatch(t *testing.T) {
	// GIVEN a Service with LatestVersion != ApprovedVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion + "-beta"

	// WHEN SetDeployedVersion is called on it with the LatestVersion
	want := service.Status.ApprovedVersion
	service.SetDeployedVersion(service.Status.LatestVersion)

	// THEN ApprovedVersion is not reset
	got := service.Status.ApprovedVersion
	if got != want {
		t.Errorf("ApprovedVersion shouldn't have changed and should still be %q, not %q",
			want, got)
	}
}

func TestSetDeployedVersionDidResetCommandFails(t *testing.T) {
	// GIVEN a Service with Commands that failed
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)
	didFail := false
	fail := true
	service.CommandController.Failed[0] = &didFail
	service.CommandController.Failed[1] = &fail
	service.CommandController.Failed[2] = &fail
	service.CommandController.Failed[3] = &didFail

	// WHEN SetDeployedVersion is called on it
	service.SetDeployedVersion(service.Status.LatestVersion)

	// THEN all the Commands Failed's become nil
	for _, failed := range service.CommandController.Failed {
		if failed != nil {
			t.Errorf("CommandController.Failed should have been reset to nil and not be %t",
				*failed)
		}
	}
}

func TestSetDeployedVersionDidResetWebHookFails(t *testing.T) {
	// GIVEN a Service with WebHooks that failed
	service := testServiceGitHub()
	whPass := testWebHookSuccessful()
	whFail := testWebHookFailing()
	service.WebHook = &webhook.Slice{
		"fail1": &whFail,
		"fail2": &whFail,
		"pass1": &whPass,
		"pass2": &whPass,
	}
	didFail := false
	fail := true
	(*service.WebHook)["fail1"].Failed = &fail
	(*service.WebHook)["fail2"].Failed = &fail
	(*service.WebHook)["pass1"].Failed = &didFail
	(*service.WebHook)["pass2"].Failed = &didFail

	// WHEN SetDeployedVersion is called on it
	service.SetDeployedVersion(service.Status.LatestVersion)

	// THEN all the Commands Failed's become nil
	for index, webhook := range *service.WebHook {
		if (*webhook).Failed != nil {
			t.Errorf("WebHook[%s].Failed should have been reset to nil and not be %t",
				index, *(*webhook).Failed)
		}
	}
}

func TestSetLatestVersion(t *testing.T) {
	// GIVEN a Service and a new version
	service := testServiceGitHub()
	version := "new"

	// WHEN SetLatestVersion is called on it
	service.SetLatestVersion(version)

	// THEN LatestVersion is set to this version
	got := service.Status.LatestVersion
	if got != version {
		t.Errorf("Expected LatestVersion to be set to %q, not %q",
			version, got)
	}
}

func TestSetLatestVersionDidSetLatestVersionTimestamp(t *testing.T) {
	// GIVEN a Service and a new version
	service := testServiceGitHub()
	version := "new"

	// WHEN SetLatestVersion is called on it
	start := time.Now()
	service.SetLatestVersion(version)

	// THEN LatestVersionTimestamp is set to now in time
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("LatestVersionTimestamp was %v ago, not recent enough!",
			since)
	}
}

func TestSetLatestVersionDidResetCommandFails(t *testing.T) {
	// GIVEN a Service with Commands that failed
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)
	didFail := false
	fail := true
	service.CommandController.Failed[0] = &didFail
	service.CommandController.Failed[1] = &fail
	service.CommandController.Failed[2] = &fail
	service.CommandController.Failed[3] = &didFail

	// WHEN SetLatestVersion is called on it
	service.SetLatestVersion(service.Status.LatestVersion)

	// THEN all the Commands Failed's become nil
	for _, failed := range service.CommandController.Failed {
		if failed != nil {
			t.Errorf("CommandController.Failed should have been reset to nil and not be %t",
				*failed)
		}
	}
}

func TestSetLatestVersionDidResetWebHookFails(t *testing.T) {
	// GIVEN a Service with WebHooks that failed
	service := testServiceGitHub()
	whPass := testWebHookSuccessful()
	whFail := testWebHookFailing()
	service.WebHook = &webhook.Slice{
		"fail1": &whFail,
		"fail2": &whFail,
		"pass1": &whPass,
		"pass2": &whPass,
	}
	didFail := false
	fail := true
	(*service.WebHook)["fail1"].Failed = &fail
	(*service.WebHook)["fail2"].Failed = &fail
	(*service.WebHook)["pass1"].Failed = &didFail
	(*service.WebHook)["pass2"].Failed = &didFail

	// WHEN SetLatestVersion is called on it
	service.SetLatestVersion(service.Status.LatestVersion)

	// THEN all the Commands Failed's become nil
	for index, webhook := range *service.WebHook {
		if (*webhook).Failed != nil {
			t.Errorf("WebHook[%s].Failed should have been reset to nil and not be %t",
				index, *(*webhook).Failed)
		}
	}
}
