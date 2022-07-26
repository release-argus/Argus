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

	github_types "github.com/release-argus/Argus/service/latest_version/api_types"
	"github.com/release-argus/Argus/utils"
)

func TestRegexCheckVersionWithNil(t *testing.T) {
	// GIVEN a Service with a nil RegexVersion
	service := testServiceGitHub()
	service.LatestVersion.Require.RegexVersion = nil

	// WHEN RegexCheckVersion is called on it
	err := service.LatestVersion.Require.RegexCheckVersion("1.2.3", jLog, utils.LogFrom{})

	// THEN RegexVersion matches as it doesn't exist, so err is nil
	var want error
	if err != nil {
		t.Errorf("err should be %v, not %q",
			want, err.Error())
	}
}

func TestRegexCheckVersionWithMatch(t *testing.T) {
	// GIVEN a Service with a RegexVersion
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexVersion = "^[0-9.]+$"

	// WHEN RegexCheckVersion is called on it with a matching version
	err := service.LatestVersion.Require.RegexCheckVersion("1.2.3", jLog, utils.LogFrom{})

	// THEN RegexVersion matche, so err is nil
	var want error
	if err != nil {
		t.Errorf("err should be %v, not %q",
			want, err.Error())
	}
}

func TestRegexCheckVersionWithNoMatch(t *testing.T) {
	// GIVEN a Service with a RegexVersion
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexVersion = "[0-9.]+$"

	// WHEN RegexCheckVersion is called on it with a non-matching version
	err := service.LatestVersion.Require.RegexCheckVersion("1.2.3-beta", jLog, utils.LogFrom{})

	// THEN RegexVersion doesn't match, so err is non-nil
	if err == nil {
		t.Errorf("err should be non-nil, not %v",
			err)
	}
}

func TestRegexCheckContentWithStringNil(t *testing.T) {
	// GIVEN a Service with a nil RegexContent
	service := testServiceGitHub()
	service.LatestVersion.Require.RegexContent = nil

	// WHEN RegexCheckContent is called on it
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", "something1.2.3.debsomething", jLog, utils.LogFrom{})

	// THEN RegexContent matches as it doesn't exist, so err is nil
	var want error
	if err != nil {
		t.Errorf("err should be %v, not %q",
			want, err.Error())
	}
}

func TestRegexCheckContentWithStringMatch(t *testing.T) {
	// GIVEN a Service with a RegexContent
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexContent = "{{ version }}\\.deb"

	// WHEN RegexCheckContent is called on it with a matching version
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", "something1.2.3.debsomething", jLog, utils.LogFrom{})

	// THEN RegexContent matche, so err is nil
	var want error
	if err != nil {
		t.Errorf("err should be %v, not %q",
			want, err.Error())
	}
}

func TestRegexCheckContentWithStringNoMatch(t *testing.T) {
	// GIVEN a Service with a RegexContent
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexContent = "{{ version }}\\.deb$"

	// WHEN RegexCheckContent is called on it with a non-matching version
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", "something1.2.3-beta.debsomething", jLog, utils.LogFrom{})

	// THEN RegexContent doesn't match, so err is non-nil
	if err == nil {
		t.Errorf("err should be non-nil, not %v",
			err)
	}
}

func TestRegexCheckContentWithGitHubAssetMatch(t *testing.T) {
	// GIVEN a Service with a RegexContent and GitHubAsset
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexContent = "{{ version }}\\.deb$"
	githubAsset := []github_types.Asset{
		{Name: "1.2.2", BrowserDownloadURL: "https://example.com/1.2.2.deb"},
		{Name: "1.2.3", BrowserDownloadURL: "https://example.com/1.2.3.deb"},
	}

	// WHEN RegexCheckContent is called on it with a non-matching version
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", githubAsset, jLog, utils.LogFrom{})

	// THEN RegexContent does match, so err is nil
	if err != nil {
		t.Errorf("err should be nil, not %q",
			err.Error())
	}
}

func TestRegexCheckContentWithGitHubAssetNoMatch(t *testing.T) {
	// GIVEN a Service with a RegexContent and GitHubAsset
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexContent = "{{ version }}\\.deb$"
	githubAsset := []github_types.Asset{
		{Name: "1.2.2", BrowserDownloadURL: "https://example.com/1.2.2.deb"},
		{Name: "1.2.3", BrowserDownloadURL: "https://example.com/1.2.3.exe"},
	}

	// WHEN RegexCheckContent is called on it with a non-matching version
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", githubAsset, jLog, utils.LogFrom{})

	// THEN RegexContent doesn't match, so err is bob-nil
	if err == nil {
		t.Errorf("err should be non-nil, not %v",
			err)
	}
}

func TestRegexCheckContentWithInvalidBodyType(t *testing.T) {
	// GIVEN a Service with a RegexContent and GitHubAsset
	jLog = utils.NewJLog("WARN", false)
	service := testServiceGitHub()
	*service.LatestVersion.Require.RegexContent = "{{ version }}\\.deb$"

	// WHEN RegexCheckContent is called on it with a non-valid body type
	err := service.LatestVersion.Require.RegexCheckContent("1.2.3", 123, jLog, utils.LogFrom{})

	// THEN RegexContent doesn't match, so err is bob-nil
	if err == nil {
		t.Errorf("err should be non-nil, not %v",
			err)
	}
}
