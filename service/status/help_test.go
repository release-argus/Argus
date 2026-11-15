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

package status

import (
	"fmt"
	"maps"
	"os"
	"sync"
	"testing"

	dbtype "github.com/release-argus/Argus/db/types"
	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
	"github.com/release-argus/Argus/service/dashboard"
)

var metricsMu sync.RWMutex

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

func testStatus() (status *Status) {
	var (
		announceChannel = make(chan []byte, 2)
		saveChannel     = make(chan bool, 5)
		databaseChannel = make(chan dbtype.Message, 5)
	)
	svcStatus := New(
		announceChannel, databaseChannel, saveChannel,
		"",
		"", "",
		"", "",
		"",
		&dashboard.Options{},
	)
	status = svcStatus
	status.Init(
		0, 0, 0,
		ServiceInfo{
			ID:         "test-service",
			Name:       "test-service-name",
			ServiceURL: "https://example.com/service/url",
		},
		&dashboard.Options{
			Icon:       "https://example.com/icon.png",
			IconLinkTo: "https://example.com/icon-link",
			WebURL:     "https://example.com",
			Tags:       []string{"foo", "bar"},
		},
	)

	status.SetDeployedVersion("0.0.0", "2001-01-01T01:01:01Z", false)
	status.SetLatestVersion("2.2.2", "2002-02-02T02:02:02Z", false)
	status.SetApprovedVersion("1.1.1", false)
	status.SetLastQueried("2002-02-02T00:00:00Z")

	return status
}

// failSliceValues renders fails by value, as [test.AssertSlicesEqual] compares pointers by address.
func failSliceValues(fails []*bool) []string {
	if fails == nil {
		return nil
	}
	values := make([]string, len(fails))
	for i, v := range fails {
		values[i] = test.StringifyPtr(v)
	}
	return values
}

// failValues renders fails by value, as [test.AssertMapEqual] compares pointers by address.
func failValues(fails map[string]*bool) map[string]string {
	if fails == nil {
		return nil
	}
	values := make(map[string]string, len(fails))
	for k, v := range fails {
		values[k] = test.StringifyPtr(v)
	}
	return values
}

// mergedMap returns a new map holding base overlaid with overrides.
func mergedMap[V any](base, overrides map[string]V) map[string]V {
	merged := make(map[string]V, len(base)+len(overrides))
	maps.Copy(merged, base)
	maps.Copy(merged, overrides)
	return merged
}
