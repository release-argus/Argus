// Copyright [2026] [Argus]
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

package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/fileaccess"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestLookup_Query__readsVersionFile(t *testing.T) {
	logtest.InitLog()

	tests := []struct {
		name          string
		content       string
		regex         string
		regexTemplate string
		wantVersion   string
	}{
		{
			name:        "capture group",
			content:     "productversion=\"7.2.2-72806\"\n",
			regex:       `productversion="([^"]+)"`,
			wantVersion: "7.2.2-72806",
		},
		{
			name:        "no regex uses file contents",
			content:     "7.2.2-72806\n",
			wantVersion: "7.2.2-72806",
		},
		{
			name:        "regex without capture group",
			content:     "release 7.2.2-72806",
			regex:       `7\.2\.2-[0-9]+`,
			wantVersion: "7.2.2-72806",
		},
		{
			name:          "regex template",
			content:       "productversion=\"v7-2-2\"\n",
			regex:         `productversion="v([0-9]+)-([0-9]+)-([0-9]+)"`,
			regexTemplate: "$1.$2.$3",
			wantVersion:   "7.2.2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookup, svcStatus := newTestLookup(t, tc.content)
			lookup.Regex = tc.regex
			lookup.RegexTemplate = tc.regexTemplate
			if err := lookup.Query(false, logx.LogFrom{}); err != nil {
				t.Fatalf("Lookup.Query() failed: %v", err)
			}

			if got := svcStatus.DeployedVersion(); got != tc.wantVersion {
				t.Errorf("DeployedVersion() = %q, want %q", got, tc.wantVersion)
			}
		})
	}
}

func TestLookup_Query__rejectsUnusableVersionMatch(t *testing.T) {
	logtest.InitLog()

	tests := []struct {
		name    string
		content string
		regex   string
		wantErr string
	}{
		{
			name:    "no match",
			content: "no version here\n",
			regex:   `productversion="([^"]+)"`,
			wantErr: "didn't return any matches",
		},
		{
			name:    "empty version",
			content: "productversion=\"\"\n",
			regex:   `productversion="([^"]*)"`,
			wantErr: "produced an empty version",
		},
		{
			name:    "empty file without regex",
			content: "",
			regex:   "",
			wantErr: "no version found",
		},
		{
			name:    "whitespace-only file without regex",
			content: " \r\n\t",
			regex:   "",
			wantErr: "no version found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookup, svcStatus := newTestLookup(t, tc.content)
			lookup.Regex = tc.regex
			err := lookup.Query(false, logx.LogFrom{})
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Lookup.Query() error = %v, want error containing %q", err, tc.wantErr)
			}
			if got := svcStatus.DeployedVersion(); got != "" {
				t.Errorf("DeployedVersion() = %q after failed query, want empty", got)
			}
		})
	}
}

func TestLookup_Decode__regexTemplate(t *testing.T) {
	lookup, err := Decode(
		"yaml",
		[]byte("path: /VERSION\nregex: 'v([0-9]+)-([0-9]+)'\nregex_template: '$1.$2'\n"),
		nil,
		nil,
		base.DefaultsConfig{},
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}
	if lookup.RegexTemplate != "$1.$2" {
		t.Errorf("RegexTemplate = %q, want %q", lookup.RegexTemplate, "$1.$2")
	}
}

func newTestLookup(t *testing.T, content string) (*Lookup, *status.Status) {
	t.Helper()

	fixturePath := filepath.Join(t.TempDir(), "VERSION")
	if err := os.WriteFile(fixturePath, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}
	if err := fileaccess.ConfigureAllowedPaths([]string{fixturePath}); err != nil {
		t.Fatalf("fileaccess.ConfigureAllowedPaths() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := fileaccess.ConfigureAllowedPaths(nil); err != nil {
			t.Errorf("fileaccess.ConfigureAllowedPaths(nil) failed: %v", err)
		}
	})

	options, err := opt.Decode(
		"yaml",
		[]byte("semantic_versioning: true"),
		opttest.PlainDefaultsConfig(),
	)
	if err != nil {
		t.Fatalf("opt.Decode() failed: %v", err)
	}
	svcStatus, err := statustest.New("yaml", nil)
	if err != nil {
		t.Fatalf("statustest.New() failed: %v", err)
	}
	svcStatus.ServiceInfo.ID = "file-test"

	lookup := &Lookup{
		Path:  fixturePath,
		Regex: `productversion="([^"]+)"`,
	}
	lookup.Init(options, svcStatus, base.DefaultsConfig{})
	return lookup, svcStatus
}
