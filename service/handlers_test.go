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

	command "github.com/release-argus/Argus/commands"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/webhook"
)

func TestUpdateLatestApprovedDidUpdateApprovedVersion(t *testing.T) {
	// GIVEN a Service whos ApprovedVersion != LatestVersion
	service := testServiceGitHub()

	// WHEN UpdateLatestApproved is called on it
	want := service.Status.LatestVersion
	service.UpdateLatestApproved()

	// THEN ApprovedVersion becomes LatestVersion
	got := service.Status.ApprovedVersion
	if got != want {
		t.Errorf("LatestVersion should have changed to %q. Got %q",
			want, got)
	}
}

func TestUpdateLatestApprovedDidAnnounce(t *testing.T) {
	// GIVEN a Service whos ApprovedVersion != LatestVersion
	service := testServiceGitHub()

	// WHEN UpdateLatestApproved is called on it
	service.UpdateLatestApproved()

	// THEN the approval appears in Announce
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdateLatestApprovedDidntAnnounceIfAlreadyApproved(t *testing.T) {
	// GIVEN a Service whos ApprovedVersion == LatestVersion
	service := testServiceGitHub()
	service.Status.ApprovedVersion = service.Status.LatestVersion

	// WHEN UpdateLatestApproved is called on it
	service.UpdateLatestApproved()

	// THEN nothing is sent to Announce
	got := len(*service.Announce)
	want := 0
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdatedVersionDidAnnounce(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN the update appears in Announce
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdatedVersionDidCheckWebHookFails(t *testing.T) {
	// GIVEN a Service with a WebHook that failed
	service := testServiceGitHub()
	failed := true
	service.WebHook = &webhook.Slice{
		"test": &webhook.WebHook{
			Failed: &failed,
		},
	}

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN no update appears in Announce
	got := len(*service.Announce)
	want := 0
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdatedVersionDidCheckCommandFails(t *testing.T) {
	// GIVEN a Service with a Command that failed
	service := testServiceGitHub()
	failed := true
	service.Command = &command.Slice{
		command.Command{},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(nil, &serviceID, nil, service.Command, nil)
	service.CommandController.Failed[0] = &failed

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN no update appears in Announce
	got := len(*service.Announce)
	want := 0
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdatedVersionDidUpdateDeployedVersion(t *testing.T) {
	// GIVEN a Service with no command/webhook fails and no DeployedVersionLookup
	service := testServiceGitHub()

	// WHEN UpdatedVersion is called on it
	want := service.Status.LatestVersion
	service.UpdatedVersion()

	// THEN the DeployedVersion is updated to LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestUpdatedVersionDidntUpdateDeployedVersion(t *testing.T) {
	// GIVEN a Service with a DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &DeployedVersionLookup{}

	// WHEN UpdatedVersion is called on it
	want := service.Status.DeployedVersion
	service.UpdatedVersion()

	// THEN the DeployedVersion stays the same
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestUpdatedVersionDidUpdateApprovedVersionIfActionsPassed(t *testing.T) {
	// GIVEN a Service with a DeployedVersionLookup
	service := testServiceGitHub()
	service.DeployedVersionLookup = &DeployedVersionLookup{}
	service.WebHook = &webhook.Slice{
		"test": &webhook.WebHook{},
	}
	failed := false
	(*service.WebHook)["test"].Failed = &failed
	service.Command = &command.Slice{
		command.Command{},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(nil, &serviceID, nil, service.Command, nil)
	service.CommandController.Failed[0] = &failed

	// WHEN UpdatedVersion is called on it
	want := service.Status.DeployedVersion
	service.UpdatedVersion()

	// THEN the DeployedVersion stays the same
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestUpdatedVersionDidTrySave(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN a save is requested from the SaveChannel
	got := len(*service.SaveChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestHandleUpdateActionsWithNothing(t *testing.T) {
	// GIVEN a Service with no Commands or WebHooks
	service := testServiceGitHub()
	service.Command = nil
	service.WebHook = nil

	// WHEN HandleUpdateActions is called on it
	want := service.Status.LatestVersion
	service.HandleUpdateActions()

	// THEN DeployedVersion is now LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have been updated to %q, not %q",
			want, got)
	}
}

func TestHandleUpdateActionsWithSuccessfulCommandAndNoAutoApprove(t *testing.T) {
	// GIVEN a Service with a Command that'll pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.DeployedVersion
	service.HandleUpdateActions()

	// THEN DeployedVersion is unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have been updated from %q to %q",
			want, got)
	}
}

func TestHandleUpdateActionsWithSuccessfulCommandAndAutoApprove(t *testing.T) {
	// GIVEN a Service with a Command that'll pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.AutoApprove = true
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.LatestVersion
	service.HandleUpdateActions()

	// THEN DeployedVersion is now LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have been updated to %q, not %q",
			want, got)
	}
}

func TestHandleUpdateActionsWithFailingCommandAndAutoApprove(t *testing.T) {
	// GIVEN a Service with a Command that'll fail
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.AutoApprove = true
	service.Command = &command.Slice{
		command.Command{"ls", "-la", "/root"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.DeployedVersion
	service.HandleUpdateActions()

	// THEN DeployedVersion is unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have been updated from %q to %q",
			want, got)
	}
}

func TestHandleFailedActionsWithNothing(t *testing.T) {
	// GIVEN a Service with no Commands or WebHooks
	service := testServiceGitHub()
	service.Command = nil
	service.WebHook = nil

	// WHEN HandleFailedActions is called on it
	want := service.Status.LatestVersion
	service.HandleFailedActions()

	// THEN DeployedVersion is now LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have been updated to %q, not %q",
			want, got)
	}
}

func TestHandleFailedActionsWithSuccessfulCommand(t *testing.T) {
	// GIVEN a Service with all Commands that'll pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleFailedActions is called on it
	want := service.Status.LatestVersion
	service.HandleFailedActions()

	// THEN DeployedVersion is now LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have been updated from %q to %q",
			want, got)
	}
}

func TestHandleFailedActionsWithFailingCommand(t *testing.T) {
	// GIVEN a Service with a Command that did fail and will now pass
	jLog = utils.NewJLog("WARN", false)
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

	// WHEN HandleFailedActions is called on it
	want := service.Status.DeployedVersion
	service.HandleFailedActions()

	// THEN DeployedVersion is unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have been updated from %q to %q",
			want, got)
	}
}

func TestHandleFailedActionsWithFailingWebHook(t *testing.T) {
	// GIVEN a Service with a WebHook that did fail and will now pass
	jLog = utils.NewJLog("WARN", false)
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
	service.WebHook.Init(
		jLog,
		service.ID,
		service.Status,
		&webhook.Slice{},
		&webhook.WebHook{},
		&webhook.WebHook{},
		nil)

	// WHEN HandleFailedActions is called on it
	want := service.Status.DeployedVersion
	service.HandleFailedActions()

	// THEN DeployedVersion is unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have been updated from %q to %q",
			want, got)
	}
}

func TestHandleFailedActionsDidOnlyRedoFailed(t *testing.T) {
	// GIVEN a Service with a Command that failed that'll pass and the ones that passed will fail
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la", "/root"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)
	passes := false
	service.CommandController.Failed[0] = &passes
	service.CommandController.Failed[2] = &passes
	failed := true
	service.CommandController.Failed[1] = &failed

	// WHEN HandleFailedActions is called on it
	want := service.Status.LatestVersion
	service.HandleFailedActions()

	// THEN DeployedVersion is unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have been updated from %q to %q",
			got, want)
	}
}

func TestHandleCommandWithNilCommands(t *testing.T) {
	// GIVEN a Service with no Commands
	service := testServiceGitHub()
	service.Command = nil

	// WHEN HandleCommand is called
	want := service.Status.DeployedVersion
	service.HandleCommand("")

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleCommandWithUnknownCommand(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(nil, &serviceID, nil, service.Command, nil)

	// WHEN HandleCommand is called for an unknown command
	want := service.Status.DeployedVersion
	service.HandleCommand("ls -la /root")

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleCommandWithSuccessfulCommand(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-lah"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleCommand is called
	want := service.Status.LatestVersion
	service.HandleCommand((*service.Command)[0].String())

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have changed from %s to %s",
			got, want)
	}
}

func TestHandleCommandWithFailingCommand(t *testing.T) {
	// GIVEN a Service with a WebHook that'll fail
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = &command.Slice{
		command.Command{"ls", "-lah", "/root"},
	}
	service.CommandController = &command.Controller{}
	serviceID := "test"
	service.CommandController.Init(jLog, &serviceID, nil, service.Command, nil)

	// WHEN HandleCommand is called
	want := service.Status.LatestVersion
	service.HandleCommand((*service.Command)[0].String())

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got == want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleWebHookWithUnknownWebHook(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	service := testServiceGitHub()
	wh := testWebHookSuccessful()
	service.WebHook = &webhook.Slice{
		"test": &wh,
	}

	// WHEN HandleWebHook is called for WebHook that doesn't exist
	want := service.Status.DeployedVersion
	service.HandleWebHook("something")

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed to %s, want %s",
			got, want)
	}
}

func TestHandleWebHookWithSuccessfulWebHook(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	service := testServiceGitHub()
	wh := testWebHookSuccessful()
	var logInitSlice webhook.Slice
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil)
	service.WebHook = &webhook.Slice{
		"test": &wh,
	}

	// WHEN HandleWebHook is called for WebHook that doesn't exist
	want := service.Status.LatestVersion
	service.HandleWebHook("test")

	// THEN the DeployedVersion will now be LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion should have changed to %s, not %s",
			want, got)
	}
}

func TestHandleWebHookWithFailingWebHook(t *testing.T) {
	// GIVEN a Service with a WebHook that'll fail
	service := testServiceGitHub()
	wh := testWebHookFailing()
	var logInitSlice webhook.Slice
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil)
	service.WebHook = &webhook.Slice{
		"test": &wh,
	}

	// WHEN HandleWebHook is called for WebHook that doesn't exist
	want := service.Status.DeployedVersion
	service.HandleWebHook("test")

	// THEN the DeployedVersion will be unchanged
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleSkipWithNotLatestVersion(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN HandleSkip is called without LatestVersion
	want := service.Status.ApprovedVersion
	service.HandleSkip("something")

	// THEN ApprovedVersion doesn't change
	got := service.Status.ApprovedVersion
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestHandleSkipDidSkip(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN HandleSkip is called with LatestVersion
	want := "SKIP_" + service.Status.LatestVersion
	service.HandleSkip(service.Status.LatestVersion)

	// THEN ApprovedVersion is set to SKIP_LatestVersion
	got := service.Status.ApprovedVersion
	if got != want {
		t.Errorf("Got %s, want %s",
			got, want)
	}
}

func TestHandleSkipDidAnnounce(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN HandleSkip is called with LatestVersion
	service.HandleSkip(service.Status.LatestVersion)

	// THEN this skip is sent to Announce
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestHandleSkipDidTrySave(t *testing.T) {
	// GIVEN a Service
	service := testServiceGitHub()

	// WHEN HandleSkip is called with LatestVersion
	service.HandleSkip(service.Status.LatestVersion)

	// THEN a save is requested from the SaveChannel
	got := len(*service.SaveChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}
