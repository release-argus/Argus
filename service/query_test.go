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
	"strings"
	"testing"

	"github.com/release-argus/Argus/service/latest_version/filters"
	"github.com/release-argus/Argus/utils"
)

func TestServiceGetVersionsWithGitHubRateLimit(t *testing.T) {
	// GIVEN a Service with a Query body erroring about reaching the GitHub rate limit
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := "something rate limit something"

	// WHEN GetVersions is called on this body
	_, err := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	_, err := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	_, err := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	_, err := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+"
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+"
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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

func TestServiceGetVersionsWithFailingURLCommand(t *testing.T) {
	// GIVEN a Service and a Query body containing a beta version
	jLog = utils.NewJLog("WARN", false)
	service := testServiceURL()
	service.URLCommands = &filters.URLCommandSlice{
		{Type: "regex", Regex: stringPtr("(argus-[0-9]+)\"")}}
	body :=
		`
		new release: "https://example.com/argus-0.1.0-beta"
		`

	// WHEN GetVersions is called on this body
	_, err := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the releases are ordered
	e := utils.ErrorToString(err)
	if !(strings.HasPrefix(e, "regex") && strings.HasSuffix(e, "didn't return any matches")) {
		t.Errorf("Should have failed url_command %q regex, not %q",
			*(*service.URLCommands)[0].Regex, e)
	}
}

func TestServiceGetVersionsWithURLCommands(t *testing.T) {
	// GIVEN a Service with URLCommand(s) to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\.0$"
	service.URLCommands = &filters.URLCommandSlice{urlCommand}
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	wantFiltered := "0.6.1"
	if len(versions) != 1 || versions[0].TagName == wantFiltered {
		t.Errorf("URLCommands with Regex % should have only filtered out the %q release from versions, got %v",
			*urlCommand.Regex, wantFiltered, versions)
	}
}

func TestServiceGetVersionWithGetVersionsFail(t *testing.T) {
	// GIVEN a Service with a Query body erroring about reaching the GitHub rate limit
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	body := "something rate limit something"

	// WHEN GetVersions is called on this body
	_, err := service.LatestVersion.GetVersion([]byte(body), utils.LogFrom{})

	// THEN err is non-nil about this rate limit being reached
	e := utils.ErrorToString(err)
	if !strings.Contains(e, "rate limit ") {
		t.Errorf("%q should've errored about reaching the rate limit, not\n%s",
			body, e)
	}
}

func TestServiceGetVersionWithRegexVersion(t *testing.T) {
	// GIVEN a Service with RegexVersion to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+$"
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\.0$"
	service.URLCommands = &filters.URLCommandSlice{urlCommand}
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
	version, _ := service.LatestVersion.GetVersion([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	want := "0.7.0"
	if version != want {
		t.Errorf("GetVersion didn't use the RegexVersion %q to filter out releases and return %q. Instead got %q",
			*service.LatestVersion.Require.RegexVersion, want, version)
	}
}

func TestServiceGetVersionsWithRegexContent(t *testing.T) {
	// GIVEN a Service with RegexContent to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.LatestVersion.Require.RegexContent = "Argus-[0-9.]+\\.linux-amd64"
	*service.LatestVersion.Require.RegexVersion = ""
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
	version, _ := service.LatestVersion.GetVersion([]byte(body), utils.LogFrom{})

	// THEN the releases are filtered
	want := "0.6.0"
	if version != want {
		t.Errorf("GetVersion didn't use the RegexContent %q to filter out releases and return %q. Instead got %q",
			*service.LatestVersion.Require.RegexContent, want, version)
	}
}

func TestServiceGetVersionsWithNoMatchingRegexContent(t *testing.T) {
	// GIVEN a Service with RegexContent to filter versions and a Query body
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.LatestVersion.Require.RegexContent = "Argus-[0-9.]+\\.linux-amd64"
	*service.LatestVersion.Require.RegexVersion = ""
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
	_, err := service.LatestVersion.GetVersion([]byte(body), utils.LogFrom{})

	// THEN an err is occured as no releases match the RegexContent
	if err == nil {
		t.Errorf("GetVersion didn't use the RegexContent %q to filter out every release and return an err. Instead got %v",
			*service.LatestVersion.Require.RegexContent, err)
	}
}

func TestServiceGetVersionsWithNoMatchingURLCommands(t *testing.T) {
	// GIVEN a Service with URLCommand(s) that will filter out all releases
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.UsePreRelease = false
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = ""
	urlCommand := testURLCommandRegex()
	*urlCommand.Regex = "^[0-9.]+\\-beta$"
	service.URLCommands = &filters.URLCommandSlice{urlCommand}
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+"
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

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
	*service.LatestVersion.Require.RegexContent = ""
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+"
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
	versions, _ := service.LatestVersion.GetVersions([]byte(body), utils.LogFrom{})

	// THEN the non-semantic releases are filtered
	wantCount := 1
	wantVersion := "0.1.1"
	if len(versions) != wantCount || versions[0].TagName != wantVersion {
		t.Errorf("GetVersions should have removed the non-semantic versions, got\n%v",
			versions)
	}
}
