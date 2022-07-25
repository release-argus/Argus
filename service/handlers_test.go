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
	"github.com/release-argus/Argus/service/deployed_version"
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
	got := len(*service.Status.AnnounceChannel)
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
	got := len(*service.Status.AnnounceChannel)
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
	got := len(*service.Status.AnnounceChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestUpdatedVersionDidCheckWebHookFails(t *testing.T) {
	// GIVEN a Service with a WebHook that failed
	service := testServiceGitHub()
	service.WebHook = webhook.Slice{
		"test": &webhook.WebHook{
			Failed: &map[string]*bool{"test": boolPtr(true)},
		},
	}

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN no update appears in Announce
	got := len(*service.Status.AnnounceChannel)
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
	service.Command = command.Slice{
		command.Command{},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(nil, &service.ID, nil, &service.Command, nil, service.Interval)
	(*service.CommandController.Failed)[0] = &failed

	// WHEN UpdatedVersion is called on it
	service.UpdatedVersion()

	// THEN no update appears in Announce
	got := len(*service.Status.AnnounceChannel)
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
	service.DeployedVersionLookup = &deployed_version.Lookup{}

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
	service.DeployedVersionLookup = &deployed_version.Lookup{}
	service.WebHook = webhook.Slice{
		"test": &webhook.WebHook{Failed: &map[string]*bool{"test": boolPtr(false)}},
	}
	service.Command = command.Slice{command.Command{}}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(nil, &service.ID, nil, &service.Command, nil, service.Interval)
	(*service.CommandController.Failed)[0] = boolPtr(false)

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

func TestHandleUpdateActionsWithNothing(t *testing.T) {
	// GIVEN a Service with no Commands or WebHooks
	service := testServiceGitHub()
	service.Command = nil
	service.WebHook = nil

	// WHEN HandleUpdateActions is called on it
	want := service.Status.LatestVersion
	service.HandleUpdateActions()
	time.Sleep(time.Second)

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
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.DeployedVersion
	service.HandleUpdateActions()
	time.Sleep(time.Second)

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
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.LatestVersion
	service.HandleUpdateActions()
	time.Sleep(time.Second)

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
	service.Command = command.Slice{
		command.Command{"ls", "-la", "/root"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)

	// WHEN HandleUpdateActions is called on it
	want := service.Status.DeployedVersion
	service.HandleUpdateActions()
	time.Sleep(time.Second)

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
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)

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

func TestHandleFailedActionsWithCommandBeforeNextRunnable(t *testing.T) {
	// GIVEN a Service with all Commands that'll pass
	// and they all failed last time
	// and one of the NextRunnable's hasn't been reached
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)
	(*service.CommandController.Failed)[0] = boolPtr(true)
	(*service.CommandController.Failed)[1] = boolPtr(true)
	nextRunnable0 := time.Now().UTC().Add(-time.Hour)
	service.CommandController.NextRunnable[0] = nextRunnable0
	nextRunnable1 := time.Now().UTC().Add(time.Hour)
	service.CommandController.NextRunnable[1] = nextRunnable1

	// WHEN HandleFailedActions is called on it
	ranAt := time.Now().UTC()
	service.HandleFailedActions()

	// THEN only the Command that's past its NextRunnable will have ran
	if utils.EvalNilPtr((*service.CommandController.Failed)[0], true) != false {
		got := "true"
		if (*service.CommandController.Failed)[0] == nil {
			got = "nil"
		}
		t.Errorf("Command 0 %q should have ran and got failed=false, not %s as it's NextRunnable was %s and it was ran at %s",
			(*service.CommandController.Command)[0].String(), got, nextRunnable0, ranAt)
	}
	if utils.EvalNilPtr((*service.CommandController.Failed)[1], false) != true {
		got := "false"
		if (*service.CommandController.Failed)[1] == nil {
			got = "nil"
		}
		t.Errorf("Command 1 %q shouldn't have ran and stayed failed=false, not %s as it's NextRunnable was %s and it was ran at %s",
			(*service.CommandController.Command)[1].String(), got, nextRunnable1, ranAt)
	}
}

func TestHandleFailedActionsWithFailingCommand(t *testing.T) {
	// GIVEN a Service with a Command that did fail and will now pass
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)
	(*service.CommandController.Failed)[0] = boolPtr(false)
	(*service.CommandController.Failed)[1] = boolPtr(true)
	(*service.CommandController.Failed)[2] = boolPtr(false)
	(*service.CommandController.Failed)[3] = boolPtr(true)

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
	service.WebHook = webhook.Slice{
		"fail0": &whFail,
		"pass0": &whPass,
	}
	service.Status.Fails.WebHook = map[string]*bool{"fail0": boolPtr(false), "pass0": boolPtr(false)}
	service.WebHook.Init(jLog, &service.ID, &service.Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, service.Interval)

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

func TestHandleFailedActionsWithWebHookBeforeNextRunnable(t *testing.T) {
	// GIVEN a Service with all WebHooks that'll pass
	// and they all failed last time
	// and one of the NextRunnable's hasn't been reached
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	whPass0 := testWebHookSuccessful()
	whPass1 := testWebHookSuccessful()
	service.WebHook = webhook.Slice{
		"pass0": &whPass0,
		"pass1": &whPass1,
	}
	service.Status.Fails.WebHook = map[string]*bool{"pass0": boolPtr(true), "pass1": boolPtr(true)}
	service.WebHook.Init(jLog, &service.ID, &service.Status, &webhook.Slice{}, &webhook.WebHook{}, &webhook.WebHook{}, nil, service.Interval)
	nextRunnable0 := time.Now().UTC().Add(-time.Hour)
	service.WebHook["pass0"].NextRunnable = nextRunnable0
	nextRunnable1 := time.Now().UTC().Add(time.Hour)
	service.WebHook["pass1"].NextRunnable = nextRunnable1

	// WHEN HandleFailedActions is called on it
	ranAt := time.Now().UTC()
	service.HandleFailedActions()

	// THEN only the WebHook that's past its NextRunnable will have ran
	if utils.EvalNilPtr(service.WebHook["pass0"].GetFailStatus(), true) != false {
		got := "true"
		if service.WebHook["pass0"].GetFailStatus() == nil {
			got = "nil"
		}
		t.Errorf("WebHook 0 should have ran and got failed=false, not %s as it's NextRunnable was %s and it was ran at %s",
			got, nextRunnable0, ranAt)
	}
	if utils.EvalNilPtr(service.WebHook["pass1"].GetFailStatus(), false) != true {
		got := "true"
		if service.WebHook["pass1"].GetFailStatus() == nil {
			got = "nil"
		}
		t.Errorf("WebHook 1 shouldn't have ran and stayed failed=false, not %s as it's NextRunnable was %s and it was ran at %s",
			got, nextRunnable1, ranAt)
	}
}

func TestHandleFailedActionsDidOnlyRedoFailed(t *testing.T) {
	// GIVEN a Service with a Command that failed that'll pass and the ones that passed will fail
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = command.Slice{
		command.Command{"ls", "-la", "/root"},
		command.Command{"ls", "-la"},
		command.Command{"ls", "-la", "/root"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)
	(*service.CommandController.Failed)[0] = boolPtr(false)
	(*service.CommandController.Failed)[2] = boolPtr(false)
	(*service.CommandController.Failed)[1] = boolPtr(true)

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
	service.Command = command.Slice{
		command.Command{"ls", "-la"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(nil, &service.ID, nil, &service.Command, nil, service.Interval)

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

func TestHandleCommandWithNextRunnableFail(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	// and time is before NextRunnable
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = command.Slice{
		command.Command{"ls", "-lah"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)
	nextRunnable := time.Now().UTC().Add(time.Minute)
	service.CommandController.NextRunnable[0] = nextRunnable

	// WHEN HandleCommand is called
	want := service.Status.DeployedVersion
	service.HandleCommand(service.Command[0].String())

	// THEN the DeployedVersion will be unchanged (Command shouldn't run)
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleCommandWithNextRunnablePass(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	// and time is after NextRunnable
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	service.Command = command.Slice{
		command.Command{"ls", "-lah"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)
	service.CommandController.NextRunnable[0] = time.Now().UTC()

	// WHEN HandleCommand is called
	want := service.Status.LatestVersion
	service.HandleCommand(service.Command[0].String())

	// THEN the DeployedVersion will now be LatestVersion
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
	service.Command = command.Slice{
		command.Command{"ls", "-lah", "/root"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(jLog, &service.ID, nil, &service.Command, nil, service.Interval)

	// WHEN HandleCommand is called
	want := service.Status.LatestVersion
	service.HandleCommand(service.Command[0].String())

	// THEN the DeployedVersion will now be LatestVersion
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
	service.WebHook = webhook.Slice{
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
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, nil)
	service.WebHook = webhook.Slice{
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

func TestHandleWebHookWithNextRunnablePass(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	// and time is after NextRunnable
	service := testServiceGitHub()
	wh := testWebHookSuccessful()
	wh.NextRunnable = time.Now().UTC()
	var logInitSlice webhook.Slice
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, service.Interval)
	service.WebHook = webhook.Slice{
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

func TestHandleWebHookWithNextRunnableFail(t *testing.T) {
	// GIVEN a Service with a WebHook that'll pass
	// and time is before NextRunnable
	service := testServiceGitHub()
	wh := testWebHookSuccessful()
	wh.NextRunnable = time.Now().UTC().Add(time.Minute)
	var logInitSlice webhook.Slice
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, service.Interval)
	service.WebHook = webhook.Slice{
		"test": &wh,
	}

	// WHEN HandleWebHook is called for WebHook that doesn't exist
	want := service.Status.DeployedVersion
	service.HandleWebHook("test")

	// THEN the DeployedVersion will now be LatestVersion
	got := service.Status.DeployedVersion
	if got != want {
		t.Errorf("DeployedVersion shouldn't have changed from %s to %s",
			want, got)
	}
}

func TestHandleWebHookWithFailingWebHook(t *testing.T) {
	// GIVEN a Service with a WebHook that'll fail
	service := testServiceGitHub()
	wh := testWebHookFailing()
	var logInitSlice webhook.Slice
	logInitSlice.Init(utils.NewJLog("WARN", false), nil, nil, nil, nil, nil, nil, service.Interval)
	service.WebHook = webhook.Slice{
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

	// THEN ApprovedVersion will be unchanged
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
	got := len(*service.Status.AnnounceChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
}

func TestShouldRetryAllWithWebHooksPassed(t *testing.T) {
	// GIVEN a Service with WebHooks at failed=false
	failed := map[string]*bool{"test": boolPtr(false), "other": boolPtr(false)}
	service := testServiceGitHub()
	service.WebHook = webhook.Slice{
		"test":  {Failed: &failed},
		"other": {Failed: &failed},
	}

	// WHEN shouldRetryAll is called on this Service
	got := service.shouldRetryAll()

	// THEN we got true
	want := true
	if got != want {
		t.Errorf("Expected retry to be %t, not %t as all webhooks were failed=%t",
			want, got, failed)
	}
}

func TestShouldRetryAllWithAWebHookFail(t *testing.T) {
	// GIVEN a Service with a WebHook that failed
	failed := map[string]*bool{"test": boolPtr(true), "other": boolPtr(false)}
	service := testServiceGitHub()
	service.WebHook = webhook.Slice{
		"test":  {Failed: &failed},
		"other": {Failed: &failed},
	}

	// WHEN shouldRetryAll is called on this Service
	got := service.shouldRetryAll()

	// THEN we got false
	want := false
	if got != want {
		t.Errorf("Expected retry to be %t, not %t as a webhook had failed=%#v",
			want, got, service.WebHook["test"].Failed)
	}
}

func TestShouldRetryAllWithCommandsPassed(t *testing.T) {
	// GIVEN a Service with Commands at failed=false
	service := testServiceGitHub()
	service.Command = command.Slice{
		{"test"},
		{"other"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(nil, &service.ID, nil, &service.Command, nil, nil)
	(*service.CommandController.Failed)[0] = boolPtr(false)
	(*service.CommandController.Failed)[1] = boolPtr(false)

	// WHEN shouldRetryAll is called on this Service
	got := service.shouldRetryAll()

	// THEN we got true
	want := true
	if got != want {
		t.Errorf("Expected retry to be %t, not %t as all webhooks were failed=%#v",
			want, got, service.WebHook["test"].Failed)
	}
}

func TestShouldRetryAllWithACommandFail(t *testing.T) {
	// GIVEN a Service with a Command that failed
	service := testServiceGitHub()
	service.Command = command.Slice{
		{"test"},
		{"other"},
	}
	service.CommandController = &command.Controller{}
	service.CommandController.Init(nil, &service.ID, nil, &service.Command, nil, nil)
	(*service.CommandController.Failed)[0] = boolPtr(false)
	(*service.CommandController.Failed)[1] = boolPtr(true)

	// WHEN shouldRetryAll is called on this Service
	got := service.shouldRetryAll()

	// THEN we got false
	want := false
	if got != want {
		t.Errorf("Expected retry to be %t, not %t as a webhook had failed=%#v",
			want, got, service.CommandController.Failed)
	}
}
