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
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	db_types "github.com/release-argus/Argus/db/types"
	service_status "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/utils"
)

func TestDeployedVersionLookupGetAllowInvalidCerts(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	dvl := testDeployedVersion()

	// WHEN GetAllowInvalidCerts is called on it
	got := dvl.GetAllowInvalidCerts()

	// THEN AllowInvalidCerts is returned
	want := dvl.AllowInvalidCerts
	if got != *want {
		t.Errorf("Got %t, want %t",
			got, *want)
	}
}

func TestDeployedVersionLookupQueryWithInvalidURL(t *testing.T) {
	// GIVEN a DeployedVersionLookup with an invalid URL
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "invalid://	test"

	// WHEN Query is called on it
	_, err := dvl.Query(utils.LogFrom{}, false)

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Expecting err to be non-nil because of the invalid url %q, not %v",
			dvl.URL, err)
	}
}

func TestDeployedVersionLookupQueryWithPassingJSON(t *testing.T) {
	// GIVEN a DeployedVersionLookup referencing JSON
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://api.github.com/repos/release-argus/argus/releases/latest"
	dvl.Regex = ""
	dvl.JSON = "url"

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, false)

	// THEN the Query is successful
	startswith := "https://api.github.com/repos/release-argus/Argus/releases/"
	if err != nil {
		t.Errorf("Query should have passed without err\n%s",
			err.Error())
	}
	if !strings.HasPrefix(version, startswith) {
		t.Errorf("Query got %q, want %q",
			version, startswith)
	}
}

func TestDeployedVersionLookupQueryWithFailingJSON(t *testing.T) {
	// GIVEN a DeployedVersionLookup referencing JSON
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://api.github.com/repos/release-argus/argus/releases/latest"
	dvl.Regex = ""
	dvl.JSON = "something"

	// WHEN Query is called on it
	_, err := dvl.Query(utils.LogFrom{}, false)

	// THEN err is non-nil
	if err == nil {
		t.Errorf("Query should have failed as %q JSON shouldn't map to anything",
			dvl.JSON)
	}
}

func TestDeployedVersionLookupQueryWithInvalidSourceJSON(t *testing.T) {
	// GIVEN a DeployedVersionLookup referencing JSON
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io"
	dvl.Regex = ""
	dvl.JSON = "something"

	// WHEN Query is called on it
	_, err := dvl.Query(utils.LogFrom{}, false)

	// THEN err is non-nil as URL isn't JSON
	if err == nil {
		t.Errorf("Query should have failed as %q JSON shouldn't map to anything",
			dvl.JSON)
	}
}

func TestDeployedVersionLookupQueryWithPassingRegex(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	dvl := testDeployedVersion()

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, false)

	// THEN the Query is successful
	want := "2022"
	if err != nil {
		t.Errorf("Query should have passed without err\n%s",
			err.Error())
	}
	if version != want {
		t.Errorf("Query got %q, want %q",
			version, want)
	}
}

func TestDeployedVersionLookupQueryWithFailingRegex(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.Regex = "^hello$"

	// WHEN Query is called on it
	_, err := dvl.Query(utils.LogFrom{}, false)

	// THEN err is non-nil as RegEx didn't match
	if err == nil {
		t.Errorf("Query should have failed as %q RegEx shouldn't match anything",
			dvl.Regex)
	}
}

func TestDeployedVersionLookupQueryWithInvalidSemanticVersioning(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, true)

	// THEN err is non-nil as version isn't semantic
	if err == nil {
		t.Errorf("Query should have failed as %q isn't semantic versioned",
			version)
	}
}

func TestDeployedVersionLookupQueryWithValidSemanticVersioning(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/getting-started/"
	dvl.Regex = "argus-([0-9.]+)\\."

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, true)

	// THEN Query returns the version
	want := "0.0.0"
	if err != nil {
		t.Errorf("Query should have passed without err as %q is valid semantic versioning\n%s",
			version, err.Error())
	}
	if version != want {
		t.Errorf("Query got %q, want %q",
			version, want)
	}
}

func TestDeployedVersionLookupQueryWithHeadersFail(t *testing.T) {
	// GIVEN a DeployedVersionLookup with invalid GitHub auth headers
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://api.github.com/repos/release-argus/argus/releases/latest"
	dvl.Regex = ""
	dvl.JSON = "message"
	dvl.Headers = []Header{
		{
			Key:   "Authorization",
			Value: "token ghp_FAIL",
		},
	}

	// WHEN Query is called on it
	version, _ := dvl.Query(utils.LogFrom{}, false)

	// THEN version is about "Bad credentials"
	want := "Bad credentials"
	if version != want {
		t.Errorf("Query should have failed as an invalid auth key was used. Want message=%q, not %q",
			want, version)
	}
}

func TestDeployedVersionLookupQueryWithBasicAuth(t *testing.T) {
	// GIVEN a DeployedVersionLookup with Basic Auth
	dvl := testDeployedVersion()
	dvl.BasicAuth = &BasicAuth{
		Username: "argus",
		Password: "test",
	}

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, false)

	// THEN the Query is successful
	want := "2022"
	if err != nil {
		t.Errorf("Query should have passed without err\n%s",
			err.Error())
	}
	if version != want {
		t.Errorf("Query got %q, want %q",
			version, want)
	}
}

func TestDeployedVersionLookupQueryWithAllowInvalidCerts(t *testing.T) {
	// GIVEN a DeployedVersionLookup
	dvl := testDeployedVersion()
	*dvl.AllowInvalidCerts = true

	// WHEN Query is called on it
	version, err := dvl.Query(utils.LogFrom{}, false)

	// THEN the Query is successful
	want := "2022"
	if err != nil {
		t.Errorf("Query should have passed without err\n%s",
			err.Error())
	}
	if version != want {
		t.Errorf("Query got %q, want %q",
			version, want)
	}
}

func TestDeployedVersionLookupCheckValuesWithNil(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *DeployedVersionLookup

	// WHEN CheckValues is called on it
	err := dvl.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("Got %s, want nil",
			err.Error())
	}
}

func TestDeployedVersionLookupCheckValuesPass(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *DeployedVersionLookup

	// WHEN CheckValues is called on it
	err := dvl.CheckValues("")

	// THEN err is nil
	if err != nil {
		t.Errorf("Got %s, want nil",
			err.Error())
	}
}

func TestDeployedVersionLookupCheckValuesFail(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	dvl := testDeployedVersion()
	dvl.URL = ""
	dvl.Regex = "hello[0-9"

	// WHEN CheckValues is called on it
	err := dvl.CheckValues("")

	// THEN 3 lines of err are printed
	e := utils.ErrorToString(err)
	errCount := strings.Count(e, "\\")
	wantCount := 3
	if errCount != wantCount {
		t.Errorf("%v is invalid, so should have %d errs, not %d!\nGot %s",
			dvl, wantCount, errCount, e)
	}
}

func TestDeployedVersionLookupPrintWithNil(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *DeployedVersionLookup
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called on it
	dvl.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 0
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestDeployedVersionLookupPrint(t *testing.T) {
	// GIVEN a fully defined DeployedVersionLookup
	dvl := testDeployedVersion()
	dvl.Headers = []Header{
		{
			Key:   "Authorization",
			Value: "token ghp_FAIL",
		},
	}
	dvl.BasicAuth = &BasicAuth{
		Username: "argus",
		Password: "test",
	}
	dvl.JSON = "yes"
	dvl.Regex = "also_yes"
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// WHEN CheckValues is called on it
	dvl.Print("")

	// THEN the expected number of lines are printed
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = stdout
	want := 11
	got := strings.Count(string(out), "\n")
	if got != want {
		t.Errorf("Print should have given %d lines, but gave %d\n%s", want, got, out)
	}
}

func TestDeployedVersionLookupTrackWithNil(t *testing.T) {
	// GIVEN a nil DeployedVersionLookup
	var dvl *DeployedVersionLookup

	// WHEN CheckValues is called on it
	start := time.Now().UTC()
	dvl.Track(&Service{})

	// THEN the function exits straight away
	since := time.Since(start)
	if since > time.Second {
		t.Errorf("Track on %v was %v ago. Should've exited straight away!",
			dvl, since)
	}
}

func TestDeployedVersionLookupTrackWithSuccessfulToLatestVersion(t *testing.T) {
	// GIVEN a Service with a working DeployedVersionLookup that will get a newer DeployedVersion
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/config/service/"
	dvl.Regex = "([0-9.]+)test"
	version := "0.0.0"
	serviceID := "test"
	semanticVersioning := true
	interval := "10s"
	databaseChannel := make(chan db_types.Message, 5)
	service := Service{
		ID:                    &serviceID,
		Interval:              &interval,
		DeployedVersionLookup: &dvl,
		Status: &service_status.Status{
			DeployedVersion: version,
			LatestVersion:   "1.2.3",
		},
		DatabaseChannel:    &databaseChannel,
		SemanticVersioning: &semanticVersioning,
		Defaults:           &Service{},
		HardDefaults:       &Service{},
	}

	// WHEN Track is called on this
	go service.DeployedVersionLookup.Track(&service)

	// THEN Service.Status.DeployedVersion is updated
	time.Sleep(2 * time.Second)
	got := service.Status.DeployedVersion
	if got == version {
		t.Errorf("%q RegEx on %s should have updated DeployedVersion from %q",
			dvl.Regex, dvl.URL, service.Status.DeployedVersion)
	}
}

func TestDeployedVersionLookupTrackWithSuccessfulToNonLatestVersion(t *testing.T) {
	// GIVEN a Service with a working DeployedVersionLookup that will get a newer DeployedVersion
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/config/service/"
	dvl.Regex = "([0-9.]+)test"
	version := "0.0.0"
	serviceID := "test"
	semanticVersioning := true
	interval := "10s"
	databaseChannel := make(chan db_types.Message, 5)
	service := Service{
		ID:                    &serviceID,
		Interval:              &interval,
		SemanticVersioning:    &semanticVersioning,
		DeployedVersionLookup: &dvl,
		Status: &service_status.Status{
			DeployedVersion: version,
			LatestVersion:   "1.2.4",
		},
		DatabaseChannel: &databaseChannel,
		Defaults:        &Service{},
		HardDefaults:    &Service{},
	}

	// WHEN Track is called on this
	go service.DeployedVersionLookup.Track(&service)

	// THEN Service.Status.DeployedVersion is updated
	time.Sleep(2 * time.Second)
	got := service.Status.DeployedVersion
	if got == service.Status.LatestVersion {
		t.Error("Shouldn't have got to LatestVersion")
	}
	if got == version {
		t.Errorf("%q RegEx on %s should have updated DeployedVersion from %q",
			dvl.Regex, dvl.URL, service.Status.DeployedVersion)
	}
}

func TestDeployedVersionLookupTrackWithSuccessfulToNewerLatestVersion(t *testing.T) {
	// GIVEN a Service with a working DeployedVersionLookup that will get a newer DeployedVersion
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/config/service/"
	dvl.Regex = "([0-9.]+)test"
	version := "0.0.0"
	serviceID := "test"
	semanticVersioning := true
	interval := "10s"
	oldLatestVersion := "1.2.2"
	databaseChannel := make(chan db_types.Message, 5)
	service := Service{
		ID:                    &serviceID,
		Interval:              &interval,
		SemanticVersioning:    &semanticVersioning,
		DeployedVersionLookup: &dvl,
		Status: &service_status.Status{
			DeployedVersion: version,
			LatestVersion:   oldLatestVersion,
		},
		DatabaseChannel: &databaseChannel,
		Defaults:        &Service{},
		HardDefaults:    &Service{},
	}

	// WHEN Track is called on this
	go service.DeployedVersionLookup.Track(&service)

	// THEN Service.Status.DeployedVersion is updated
	time.Sleep(2 * time.Second)
	if service.Status.DeployedVersion != service.Status.LatestVersion {
		t.Errorf("LatestVersion %q shouldn't be lower than DeployedVersion %q",
			service.Status.LatestVersion, service.Status.DeployedVersion)
	}
}

func TestDeployedVersionLookupTrackWithSuccessfulTriggersSave(t *testing.T) {
	// GIVEN a Service with a working DeployedVersionLookup that will get a newer DeployedVersion
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/config/service/"
	dvl.Regex = "([0-9.]+)test"
	version := "0.0.0"
	service := testServiceURL()
	service.DeployedVersionLookup = &dvl
	service.Status.DeployedVersion = version
	service.Status.LatestVersion = "1.2.3"

	// WHEN Track is called on this
	go service.DeployedVersionLookup.Track(&service)

	// THEN a Save is sent to the Service.SaveChannel
	time.Sleep(2 * time.Second)
	got := len(*service.SaveChannel)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the DeployedVersion change. Should be %d",
			got, want)
	}
}

func TestDeployedVersionLookupTrackWithSuccessfulTriggersWebHookAnnounce(t *testing.T) {
	// GIVEN a Service with a working DeployedVersionLookup that will get a newer DeployedVersion
	jLog = utils.NewJLog("WARN", false)
	dvl := testDeployedVersion()
	dvl.URL = "https://release-argus.io/docs/config/service/"
	dvl.Regex = "([0-9.]+)test"
	version := "0.0.0"
	serviceID := "test"
	semanticVersioning := true
	interval := "10s"
	announceChannel := make(chan []byte, 5)
	databaseChannel := make(chan db_types.Message, 5)
	service := Service{
		ID:                    &serviceID,
		Interval:              &interval,
		SemanticVersioning:    &semanticVersioning,
		DeployedVersionLookup: &dvl,
		Status: &service_status.Status{
			DeployedVersion:          version,
			DeployedVersionTimestamp: "2022-01-01T01:01:01Z",
			LatestVersion:            "1.2.3",
		},
		DatabaseChannel: &databaseChannel,
		Defaults:        &Service{},
		HardDefaults:    &Service{},
		Announce:        &announceChannel,
	}

	// WHEN Track is called on this
	go service.DeployedVersionLookup.Track(&service)

	// THEN a Save is sent to the Service.SaveChannel
	time.Sleep(2 * time.Second)
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the DeployedVersion change. Should be %d",
			got, want)
	}
}
