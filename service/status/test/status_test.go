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

package test

import (
	"fmt"
	"testing"

	"github.com/release-argus/Argus/config/decode"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestNew(t *testing.T) {
	type wants struct {
		info.ServiceInfo
		statusTimestamps
	}
	// GIVEN: arguments to create a Service with.
	tests := []struct {
		name     string
		format   string
		data     string
		errRegex string
		wants    wants
	}{
		{
			name:     "JSON/empty",
			format:   "json",
			data:     "",
			errRegex: `failed to unmarshal`,
			wants: wants{
				ServiceInfo:      info.ServiceInfo{},
				statusTimestamps: statusTimestamps{},
			},
		},
		{
			name:     "JSON/empty object",
			format:   "json",
			data:     "{}",
			errRegex: `^$`,
			wants: wants{
				ServiceInfo:      info.ServiceInfo{},
				statusTimestamps: statusTimestamps{},
			},
		},
		{
			name:     "YAML/empty",
			format:   "yaml",
			data:     "",
			errRegex: `^$`,
			wants: wants{
				ServiceInfo:      info.ServiceInfo{},
				statusTimestamps: statusTimestamps{},
			},
		},
		{
			name:   "JSON/filled",
			format: "json",
			data: test.TrimJSON(`{
				"approved_version": "a",
				"deployed_version": "d",
				"deployed_version_timestamp": "2006-01-02T15:04:05Z07:00",
				"latest_version": "l",
				"latest_version_timestamp": "2006-02-02T15:04:05Z07:00",
				"last_queried": "2006-03-02T15:04:05Z07:00"
			}`),
			wants: wants{
				ServiceInfo: info.ServiceInfo{
					ApprovedVersion: "a",
					DeployedVersion: "d",
					LatestVersion:   "l",
				},
				statusTimestamps: statusTimestamps{
					DeployedVersionTimestamp: "2006-01-02T15:04:05Z07:00",
					LatestVersionTimestamp:   "2006-02-02T15:04:05Z07:00",
					LastQueried:              "2006-03-02T15:04:05Z07:00",
				},
			},
		},
		{
			name:   "YAML/filled",
			format: "yaml",
			data: test.TrimYAML(`
				approved_version: a
				deployed_version: d
				deployed_version_timestamp: 2006-01-02T15:04:05Z07:00
				latest_version: l
				latest_version_timestamp: 2006-02-02T15:04:05Z07:00
				last_queried: 2006-03-02T15:04:05Z07:00
			`),
			wants: wants{
				ServiceInfo: info.ServiceInfo{
					ApprovedVersion: "a",
					DeployedVersion: "d",
					LatestVersion:   "l",
				},
				statusTimestamps: statusTimestamps{
					DeployedVersionTimestamp: "2006-01-02T15:04:05Z07:00",
					LatestVersionTimestamp:   "2006-02-02T15:04:05Z07:00",
					LastQueried:              "2006-03-02T15:04:05Z07:00",
				},
			},
		},
		{
			name:   "invalid info.ServiceInfo",
			format: "yaml",
			data:   `{"approved_version": "1.2.3}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal ServiceInfo:
					[^\s]+ could not find end character`),
		},
		{
			name:   "invalid statusTimestamps",
			format: "yaml",
			data:   `{"last_queried": ["1.2.3"]}`,
			errRegex: test.TrimYAML(`
				^failed to unmarshal timestamps:
					[^\s]+ (cannot|unable to) unmarshal`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// WHEN: the Status is created with them.
			result, err := New(tc.format, []byte(tc.data))

			prefix := fmt.Sprintf(
				"%s\nNew(format=%q, data=%q)",
				packageName, tc.format, tc.data,
			)

			// THEN: the error is as expected.
			e := errfmt.FormatError(err)
			if !util.RegexCheck(tc.errRegex, e) {
				t.Errorf(
					"%s error mismatch\ngot:  %q\nwant: %q",
					prefix, e, tc.errRegex,
				)
			}
			if e != "" {
				return
			}

			// AND: the values are set as expected.
			gotServiceInfo := decode.ToYAMLString(result.ServiceInfo, "")
			wantServiceInfo := decode.ToYAMLString(tc.wants.ServiceInfo, "")
			if gotServiceInfo != wantServiceInfo {
				t.Errorf(
					"%s\nstringified ServiceInfo mismatch\ngot:  %q\nwant: %q",
					prefix, gotServiceInfo, wantServiceInfo,
				)
			}
			if got := result.ApprovedVersion(); got != tc.wants.ApprovedVersion {
				t.Errorf(
					"%s\n .ApprovedVersion() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.ApprovedVersion,
				)
			}
			if got := result.DeployedVersionTimestamp(); got != tc.wants.DeployedVersionTimestamp {
				t.Errorf(
					"%s\n .DeployedVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.DeployedVersionTimestamp,
				)
			}
			if got := result.LatestVersionTimestamp(); got != tc.wants.LatestVersionTimestamp {
				t.Errorf(
					"%s .LatestVersionTimestamp() mismatch\ngot:  %q\nwant: %q",
					prefix, got, tc.wants.LatestVersionTimestamp,
				)
			}
		})
	}
}
