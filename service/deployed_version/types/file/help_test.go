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

//go:build unit || integration

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/fileaccess"
	"github.com/release-argus/Argus/internal/logx"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/deployed_version/types/base"
	opt "github.com/release-argus/Argus/service/option"
	opttest "github.com/release-argus/Argus/service/option/test"
	"github.com/release-argus/Argus/service/status"
)

// allowedDir is the single file_read_paths entry for the package's tests.
// The allowed paths are process-wide, so every test writes beneath this
// directory rather than setting its own, which lets tests run in parallel.
var allowedDir string

func TestMain(m *testing.M) {
	logtest.InitLog()

	dir, err := os.MkdirTemp("", "argus-dv-file-")
	if err != nil {
		fmt.Println("os.MkdirTemp() failed:", err)
		os.Exit(1)
	}
	allowedDir = dir
	if err := fileaccess.SetAllowedPaths([]string{allowedDir}); err != nil {
		fmt.Println("fileaccess.SetAllowedPaths() failed:", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	_ = os.RemoveAll(allowedDir)
	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// writeVersionFile writes content to a file beneath [allowedDir], and returns its path.
func writeVersionFile(t *testing.T, content string) string {
	t.Helper()

	dir, err := os.MkdirTemp(allowedDir, "")
	if err != nil {
		t.Fatalf("os.MkdirTemp() failed: %v", err)
	}
	path := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	return path
}

func testLookup(t *testing.T, content string) *Lookup {
	t.Helper()

	dvCfg := plainDefaultsConfig()
	optCfg := opttest.PlainDefaultsConfig()
	options, err := opt.Decode("yaml", []byte("semantic_versioning: true"), optCfg)
	if err != nil {
		t.Fatalf("opt.Decode() failed: %v", err)
	}
	announceChannel := make(chan []byte, 24)
	saveChannel := make(chan bool, 5)
	databaseChannel := make(chan dbtype.Message, 5)
	svcDashboard := &dashboard.Options{
		WebURL: "https://example.com",
	}
	svcStatus := status.New(
		announceChannel, databaseChannel, saveChannel,
		"",
		"", "",
		"", "",
		"",
		svcDashboard,
	)
	svcStatus.Init(
		0, 0, 0,
		status.ServiceInfo{
			ID: "file-testLookup",
		},
		svcDashboard,
	)

	lookup, err := Decode(
		"yaml", []byte("path: "+writeVersionFile(t, content)),
		options,
		svcStatus,
		dvCfg,
	)
	if err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	return lookup
}

// plainDefaultsConfig returns plain defaults and hardDefaults for testing.
func plainDefaultsConfig() base.DefaultsConfig {
	hardDefaults := &base.Defaults{Options: &opt.Defaults{}}
	hardDefaults.Default()
	hardDefaults.Options.Default()

	return base.DefaultsConfig{
		Soft: &base.Defaults{Options: &opt.Defaults{}},
		Hard: hardDefaults,
	}
}
