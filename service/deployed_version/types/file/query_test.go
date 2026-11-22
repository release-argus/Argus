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

	lookup, svcStatus := newTestLookup(t, "productversion=\"7.2.2-72806\"\n")
	if err := lookup.Query(false, logx.LogFrom{}); err != nil {
		t.Fatalf("Lookup.Query() failed: %v", err)
	}

	if got, want := svcStatus.DeployedVersion(), "7.2.2-72806"; got != want {
		t.Errorf("DeployedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}

func TestLookup_Query__rejectsUnusableVersionMatch(t *testing.T) {
	logtest.InitLog()

	tests := []struct {
		name    string
		content string
	}{
		{name: "no match", content: "no version here\n"},
		{name: "empty capture", content: "productversion=\"\"\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lookup, svcStatus := newTestLookup(t, tc.content)
			err := lookup.Query(false, logx.LogFrom{})
			if err == nil || !strings.Contains(err.Error(), "did not capture a version") {
				t.Fatalf("Lookup.Query() error = %v, want missing captured-version error", err)
			}
			if got := svcStatus.DeployedVersion(); got != "" {
				t.Errorf("DeployedVersion() = %q after failed query, want empty", got)
			}
		})
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
