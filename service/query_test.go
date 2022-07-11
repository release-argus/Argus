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
	"encoding/json"
	"strings"
	"testing"

	"github.com/release-argus/Argus/utils"
	api_types "github.com/release-argus/Argus/web/api/types"
)

func TestServiceHTTPRequestWithInvalidURL(t *testing.T) {
	// GIVEN a Service with an invalid URL
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "invalid://	test"

	// WHEN Query is called on this Service
	_, err := service.httpRequest(utils.LogFrom{})

	// THEN err is non-nil because of the invalid URL
	if err == nil {
		t.Errorf("err should be non-nil because of the invalid url %q",
			*service.URL)
	}
}

func TestServiceHTTPRequestWithValidURL(t *testing.T) {
	// GIVEN a DeployedVersionLookup referencing JSON
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io"
	*service.AllowInvalidCerts = true

	// WHEN Query is called on it
	_, err := service.httpRequest(utils.LogFrom{})

	// THEN err is non-nil as URL isn't JSON
	if err != nil {
		t.Errorf("Query should passed, not\n%s",
			err.Error())
	}
}

func TestServiceHTTPRequestWithUnknownURL(t *testing.T) {
	// GIVEN a DeployedVersionLookup referencing JSON
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.invalid-tld"
	*service.AllowInvalidCerts = true

	// WHEN Query is called on it
	_, err := service.httpRequest(utils.LogFrom{})

	// THEN err is nil as URL shou;dn't resolve to anything
	if err == nil {
		t.Errorf("Query should fail at lookup up %q",
			*service.URL)
	}
}

func TestServiceQueryWithInvalidURL(t *testing.T) {
	// GIVEN a Service with an invalid URL
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "invalid://	test"

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil because of the invalid URL
	if err == nil {
		t.Errorf("err should be non-nil because of the invalid url %q",
			*service.URL)
	}
}

func TestServiceQueryWithNonSemanticVersion(t *testing.T) {
	// GIVEN a Service with URLCommands that won't match
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.SemanticVersioning = true
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil because of the non-semantic version returned
	contains := "semantic version"
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contains) {
		t.Errorf("err should be non-nil because of the non-semantic version returned, not\n%s",
			e)
	}
}

func TestServiceQueryWithNewVersionRegexContentFail(t *testing.T) {
	// GIVEN a Service with RegexContent that won't match
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io"
	*service.RegexContent = "argus[0-9]+.exe"
	*service.RegexVersion = ""
	*service.SemanticVersioning = false

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil because of the invalid URL
	contains := " not matched on content for version "
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contains) {
		t.Errorf("err should be non-nil because of the invalid regex_content %q, not\n%s",
			*service.RegexContent, e)
	}
}

func TestServiceQueryWithNewVersionRegexVersionFail(t *testing.T) {
	// GIVEN a Service with RegexVersion that won't match
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io"
	*service.RegexContent = ""
	*service.RegexVersion = "^[0-9.]+$"
	*service.SemanticVersioning = false

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil because of the invalid URL
	contains := " not matched on version "
	e := utils.ErrorToString(err)
	if !strings.Contains(e, contains) {
		t.Errorf("err should be non-nil because of the invalid regex_version %q, not\n%s",
			*service.RegexVersion, e)
	}
}

func TestServiceQueryWithValidSemanticNewVersion(t *testing.T) {
	// GIVEN a Service
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = "0.0.0"

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is nil as the Query was successful
	if err != nil {
		t.Errorf("Query should have passed, not\n%s",
			err.Error())
	}
}

func TestServiceQueryWithInvalidSemanticNewVersion(t *testing.T) {
	// GIVEN a Service which will return a non-semantic version
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = "0.0.0"

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil as the Query should've returned a non-semanticVersion
	if err == nil {
		t.Errorf("Query should have failed as the %q RegEx isn't semantic versioning, not\n%s",
			*(*service.URLCommands)[0].Regex, err.Error())
	}
}

func TestServiceQueryWithInvalidSemanticOldVersion(t *testing.T) {
	// GIVEN a Service with a non-semantic LatestVersion
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = "0"

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil as the Query should've returned a non-semanticVersion
	if err == nil {
		t.Errorf("Query should have failed as the %q RegEx isn't semantic versioning, not\n%s",
			*(*service.URLCommands)[0].Regex, err.Error())
	}
}

func TestServiceQueryWithOlderNewVersion(t *testing.T) {
	// GIVEN a Service with LatestVersion newer than Query will return
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = "1.2.4"

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil as the Query should've returned a non-semanticVersion
	if err == nil {
		t.Errorf("Query should have failed as the version returned should've been lower than %q",
			service.Status.LatestVersion)
	}
}

func TestServiceQueryWithInvalidSemanticFirstVersion(t *testing.T) {
	// GIVEN a Service with a non-semantic LatestVersion
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = ""

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is non-nil as the Query should've returned a non-semanticVersion
	if err == nil {
		t.Errorf("Query should have failed as the %q RegEx isn't semantic versioning, not\n%s",
			*(*service.URLCommands)[0].Regex, err.Error())
	}
}

func TestServiceQueryWithValidFirstVersion(t *testing.T) {
	// GIVEN a Service with no FirstVersion
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = ""

	// WHEN Query is called on this Service
	_, err := service.Query()

	// THEN err is nil as the Query was successful
	if err != nil {
		t.Errorf("Query should have passed, not\n%s",
			err.Error())
	}
}

func TestServiceQuerySameVersion(t *testing.T) {
	// GIVEN a Service with no FirstVersion
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = ""

	// WHEN Query is called on this Service twice
	service.Query()
	<-*service.Announce
	service.Query()

	// THEN the function announces the query to the channel
	got := len(*service.Announce)
	want := 1
	if got != want {
		t.Errorf("%d messages in the channel from the announce. Should be %d",
			got, want)
	}
	var data api_types.WebSocketMessage
	msg := <-*service.Announce
	json.Unmarshal(msg, &data)
	if data.ServiceData.Status.LastQueried == "" {
		t.Errorf("expecting query to be announced to the websocket, not\n%v",
			data)
	}
}

func TestServiceQueryFirstVersionWithNoDeployedLookup(t *testing.T) {
	// GIVEN a Service with no DeployedVersionLookup or Actions
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	*service.URL = "https://release-argus.io/docs/config/service/"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "([0-9.]+)test"
	urlCommand.Index = 0
	service.URLCommands = &URLCommandSlice{
		urlCommand,
	}
	*service.RegexContent = ""
	*service.RegexVersion = ""
	service.Status.LatestVersion = ""
	service.Status.DeployedVersion = ""

	// WHEN Query is called on this Service
	was := service.Status.DeployedVersion
	service.Query()

	// THEN the DeployedVersion is updated
	got := service.Status.DeployedVersion
	if got == was {
		t.Errorf("DeployedVersion should've changed from %q",
			got)
	}
}

func TestServiceGetVersionsWithGitHubRateLimit(t *testing.T) {
	// GIVEN a Service with a Query body erroring about reaching the GitHub rate limit
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := "something rate limit something"

	// WHEN GetVersions is called on this body
	_, err := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN err is non-nil about this rate limit being reached
	e := utils.ErrorToString(err)
	if !strings.Contains(e, "rate limit ") {
		t.Errorf("%q should've errored about reaching the rate limit, not\n%s",
			body, e)
	}
}

func TestServiceGetVersionsWithGitHubBadCredentials(t *testing.T) {
	// GIVEN a Service with a Query body erroring about bad credentials being used
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := "something Bad credentials something"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN GetVersions is called on this body
	_, err := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN err is non-nil about these bad credentials
	e := utils.ErrorToString(err)
	if !strings.Contains(e, " access token is invalid") {
		t.Errorf("%q should've errored about credentials being bad, not\n%s",
			body, e)
	}
}

func TestServiceGetVersionsWithGitHubNoTagNames(t *testing.T) {
	// GIVEN a Service with a Query body erroring
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := "something something something"
	// Switch Fatal to panic and disable this panic.
	jLog.Testing = true
	defer func() { _ = recover() }()

	// WHEN GetVersions is called on this body
	_, err := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN err is non-nil about tag_name not being found
	e := utils.ErrorToString(err)
	if !strings.Contains(e, "tag_name ") {
		t.Errorf("%q should've errored about tag_name not being found, not\n%s",
			body, e)
	}
}

func TestServiceGetVersionsWithGitHubInvalidData(t *testing.T) {
	// GIVEN a Service with a Query body that doesn't match the defined type
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := `"tag_name":"argus"\n"url":"test"\n"url":"another_url"`

	// WHEN GetVersions is called on this body
	_, err := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN err is non-nil about the unmarshal failing
	e := utils.ErrorToString(err)
	if !strings.Contains(e, "unmarshal ") {
		t.Errorf("%q should've errored about unmarshal failing, not\n%s",
			body, e)
	}
}

func TestServiceGetVersionsWithGitHubFilterPreReleases(t *testing.T) {
	// GIVEN a Service that doesn't accept prereleases with a Query body containing a prerelease
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+"
	body :=
		`[
			{
			  "tag_name": "0.6.1",
			  "prerelease": false
			},
			{
			  "tag_name": "0.7.0",
			  "prerelease": true
			}
		]`
	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN all prereleases are removed
	if len(versions) == 0 {
		t.Error("GetVersions shouldn't have filtered out the non PreReleases")
	}
	for _, version := range versions {
		if version.PreRelease {
			t.Errorf("GetVersions should have filtered out all PreReleases since Service.UsePreRelease is %t, got\n%v",
				*service.UsePreRelease, version)
		}
	}
}

func TestServiceGetVersionsWithGitHubSortReleases(t *testing.T) {
	// GIVEN a Service and a Query body containing a patch for an older minor version higher on the relases list
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+"
	*service.SemanticVersioning = true
	body :=
		`[
			{
			  "tag_name": "0.6.3"
			},
			{
			  "tag_name": "0.8.0"
			},
			{
			  "tag_name": "0.6.2"
			},
			{
			  "tag_name": "0.7.0"
			},
			{
			  "tag_name": "0.6.1"
			}
		]`

	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the releases are ordered
	wantedOrder := []string{"0.8.0", "0.7.0", "0.6.3", "0.6.2", "0.6.1"}
	if len(versions) != len(wantedOrder) {
		t.Error("GetVersions shouldn't have filtered out any releases")
	}
	for i, version := range wantedOrder {
		if versions[i].TagName != version {
			t.Errorf("GetVersions should have sorted the releases so that ordering is %v, got\n%v",
				wantedOrder, versions)
		}
	}
}

func TestServiceGetVersionsWithURLCommands(t *testing.T) {
	// GIVEN a Service with URLCommand(s) to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\.0$"
	service.URLCommands = &URLCommandSlice{urlCommand}
	body :=
		`[
			{
			  "tag_name": "0.6.1"
			},
			{
			  "tag_name": "0.7.0"
			}
		]`

	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	wantFiltered := "0.6.1"
	if len(versions) != 1 || versions[0].TagName == wantFiltered {
		t.Errorf("URLCommands with Regex % should have only filtered out the %q release from versions, got %v",
			*urlCommand.Regex, wantFiltered, versions)
	}
}

func TestServiceGetVersionWithRegexVersion(t *testing.T) {
	// GIVEN a Service with RegexVersion to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+$"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\.0$"
	service.URLCommands = &URLCommandSlice{urlCommand}
	body :=
		`[
			{
			  "tag_name": "0.6.1-beta"
			},
			{
			  "tag_name": "0.7.0"
			}
		]`

	// WHEN GetVersions is called on this body
	version, _ := service.GetVersion([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	want := "0.7.0"
	if version != want {
		t.Errorf("GetVersion didn't use the RegexVersion %q to filter out releases and return %q. Instead got %q",
			*service.RegexVersion, want, version)
	}
}

func TestServiceGetVersionsWithRegexContent(t *testing.T) {
	// GIVEN a Service with RegexContent to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = "Argus-[0-9.]+\\.linux-amd64"
	*service.RegexVersion = ""
	body :=
		`[
			{
			  "tag_name": "0.7.0",
			  "assets": [
				{
				  "name": "Argus-0.7.0.windows-amd64",
				  "browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.7.0/Argus-0.7.0.windows-amd64"
				}
			  ]
			},
			{
			  "tag_name": "0.6.0",
			  "assets": [
				{
				  "name": "Argus-0.6.0.windows-amd64",
				  "browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.6.0/Argus-0.6.0.linux-amd64"
				},
				{
				  "name": "Argus-0.6.0.linux-amd64",
				  "browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.6.0/Argus-0.6.0.windows-amd64"
				}
			  ]
			}
		]`

	// WHEN GetVersions is called on this body
	version, _ := service.GetVersion([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	want := "0.6.0"
	if version != want {
		t.Errorf("GetVersion didn't use the RegexContent %q to filter out releases and return %q. Instead got %q",
			*service.RegexContent, want, version)
	}
}

func TestServiceGetVersionsWithNoMatchingRegexContent(t *testing.T) {
	// GIVEN a Service with RegexContent to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = "Argus-[0-9.]+\\.linux-amd64"
	*service.RegexVersion = ""
	body :=
		`[
			{
			  "tag_name": "0.7.0",
			  "assets": [
				{
				  "name": "Argus-0.7.0.windows-amd64",
				  "browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.7.0/Argus-0.7.0.windows-amd64"
				}
			  ]
			},
			{
			  "tag_name": "0.6.0",
			  "assets": [
				{
				  "name": "Argus-0.6.0.windows-amd64",
				  "browser_download_url": "https://github.com/release-argus/Argus/releases/download/0.6.0/Argus-0.6.0.windows-amd64"
				}
			  ]
			}
		]`

	// WHEN GetVersions is called on this body
	_, err := service.GetVersion([]byte(body), utils.LogFrom{})

	// THEN an err is occured as no releases match the RegexContent
	if err == nil {
		t.Errorf("GetVersion didn't use the RegexContent %q to filter out every release and return an err. Instead got %v",
			*service.RegexContent, err)
	}
}

func TestServiceGetVersionsWithNoMatchingURLCommands(t *testing.T) {
	// GIVEN a Service with URLCommand(s) that will filter out all releases
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.RegexContent = ""
	*service.RegexVersion = ""
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\-beta$"
	service.URLCommands = &URLCommandSlice{urlCommand}
	body :=
		`[
			{
			  "tag_name": "0.6.1"
			},
			{
			  "tag_name": "0.7.0"
			}
		]`

	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	if len(versions) != 0 {
		t.Errorf("URLCommands with Regex % should have filtered out all releases from versions, got %v",
			*urlCommand.Regex, versions)
	}
}

func TestServiceGetVersionsWithNonSemanticVersioning(t *testing.T) {
	// GIVEN a Service with non-semantic versioning and a Query body with non-semantic versions
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+"
	*service.SemanticVersioning = false
	body :=
		`[
			{
			  "tag_name": "07"
			},
			{
			  "tag_name": "06"
			},
			{
			  "tag_name": "0.1.1"
			}
		]`

	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN no releases are filtered
	want := 3
	if len(versions) != want {
		t.Errorf("GetVersions shouldn't have removed releases with Service.SemanticVersioning %t",
			*service.SemanticVersioning)
	}
}

func TestServiceGetVersionsWithSemanticVersioningAndSomeNonSemanticReleases(t *testing.T) {
	// GIVEN a Service wanrinf semantic versioning and a Query body with both semantic and non-semantic versions
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.RegexContent = ""
	*service.RegexVersion = "[0-9.]+"
	body :=
		`[
			{
			  "tag_name": "07"
			},
			{
			  "tag_name": "06"
			},
			{
			  "tag_name": "0.1.1"
			}
		]`

	// WHEN GetVersions is called on this body
	versions, _ := service.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the non-semantic releases are filtered
	wantCount := 1
	wantVersion := "0.1.1"
	if len(versions) != wantCount || versions[0].TagName != wantVersion {
		t.Errorf("GetVersions should have removed the non-semantic versions, got\n%v",
			versions)
	}
}
