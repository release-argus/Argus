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

package service_status

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

func TestInitWithNil(t *testing.T) {
	// GIVEN we have a Status and no shoutrrrs or webhooks
	shoutrrrs := 0
	webhooks := 0
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks)

	// THEN Fails will be empty
	if status.Fails.Shoutrrr != nil || status.Fails.WebHook != nil {
		t.Errorf("Init with %d shoutrrrs and %d webhooks should have nil Fails respectively, not %v",
			shoutrrrs, webhooks, status.Fails)
	}
}

func TestInitWithShoutrrs(t *testing.T) {
	// GIVEN we have a Status and some shoutrrs
	shoutrrrs := 4
	webhooks := 0
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks)

	// THEN Fails will be empty
	got := len(*status.Fails.Shoutrrr)
	if got != shoutrrrs {
		t.Errorf("Init with %d shoutrrrs should have made %d Fails, not %d",
			shoutrrrs, shoutrrrs, got)
	}
}

func TestInitWithWebHooks(t *testing.T) {
	// GIVEN we have a Status and some webhooks
	shoutrrrs := 0
	webhooks := 4
	var status Status

	// WHEN Init is called
	status.Init(shoutrrrs, webhooks)

	// THEN Fails will be empty
	got := len(*status.Fails.WebHook)
	if got != webhooks {
		t.Errorf("Init with %d webhooks should have made %d Fails, not %d",
			webhooks, webhooks, got)
	}
}

func TestSetLastQueried(t *testing.T) {
	// GIVEN we have a Status and some webhooks
	var status Status

	// WHEN we SetLastQueried
	start := time.Now()
	status.SetLastQueried()

	// THEN LastQueried will have been set to the current timestamp
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("LastQueried was %v ago, not recent enough!",
			since)
	}
}

func TestPrintFull(t *testing.T) {
	// GIVEN we have a Status with everything defined
	status := Status{
		ApprovedVersion:          "1.2.4",
		DeployedVersion:          "1.2.3",
		DeployedVersionTimestamp: "2022-01-01T01:01:01Z",
		LatestVersion:            "1.2.4",
		LatestVersionTimestamp:   "2022-01-01T01:01:01Z",
	}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we SetLastQueried
	status.Print("")

	// THEN a line would have been printed for each var
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 5
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestPrintEmpty(t *testing.T) {
	// GIVEN we have a Status with nothing defined
	status := Status{}
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN we SetLastQueried
	status.Print("")

	// THEN no lines would have been printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}
