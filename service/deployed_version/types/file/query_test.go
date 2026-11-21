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
	"path/filepath"
	"testing"

	"github.com/release-argus/Argus/internal/fileaccess"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	statustest "github.com/release-argus/Argus/service/status/test"
)

func TestLookup_QueryReadsSynologyVersion(t *testing.T) {
	logtest.InitLog()

	fixturePath, err := filepath.Abs(
		"../../../../web/ui/react-app/tests/fixtures/files/Synology/VERSION",
	)
	if err != nil {
		t.Fatalf("filepath.Abs() failed: %v", err)
	}
	if err := fileaccess.ConfigureAllowedRoots([]string{filepath.Dir(fixturePath)}); err != nil {
		t.Fatalf("fileaccess.ConfigureAllowedRoots() failed: %v", err)
	}
	t.Cleanup(func() {
		if err := fileaccess.ConfigureAllowedRoots(nil); err != nil {
			t.Errorf("fileaccess.ConfigureAllowedRoots(nil) failed: %v", err)
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
	if err := lookup.Query(false, logx.LogFrom{}); err != nil {
		t.Fatalf("Lookup.Query() failed: %v", err)
	}

	if got, want := svcStatus.DeployedVersion(), "7.2.2-72806"; got != want {
		t.Errorf("DeployedVersion() mismatch\ngot:  %q\nwant: %q", got, want)
	}
}
