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

package config

import (
	"slices"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/test"
)

func TestDataSettings_CheckValues(t *testing.T) {
	allowedPath := t.TempDir()
	test.IsolateEnv(t, map[string]string{"ARGUS_FILE_READ_PATH": allowedPath})

	tests := []struct {
		name    string
		paths   []string
		wantErr string
	}{
		{
			name:  "valid path with env expansion",
			paths: []string{"${ARGUS_FILE_READ_PATH}"},
		},
		{
			name:    "missing path",
			paths:   []string{allowedPath + "/missing"},
			wantErr: "file_read_paths",
		},
		{
			name:    "empty path",
			paths:   []string{""},
			wantErr: "must not contain an empty path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := DataSettings{FileReadPaths: tc.paths}
			err := settings.CheckValues()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("DataSettings.CheckValues() failed: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("DataSettings.CheckValues() error = %v, want text %q", err, tc.wantErr)
			}
		})
	}
}

func TestSettings_DataFileReadPaths(t *testing.T) {
	test.IsolateEnv(t, map[string]string{"ARGUS_FILE_READ_PATH": "/env/path"})

	settings := Settings{
		SettingsBase: SettingsBase{
			Data: DataSettings{FileReadPaths: []string{"${ARGUS_FILE_READ_PATH}", "/config/second"}},
		},
		HardDefaults: SettingsBase{
			Data: DataSettings{FileReadPaths: []string{"/hard-default"}},
		},
	}
	if got, want := settings.DataFileReadPaths(), []string{"/env/path", "/config/second"}; !slices.Equal(got, want) {
		t.Errorf("DataFileReadPaths() = %v, want %v", got, want)
	}

	settings.FromFlags.Data.FileReadPaths = []string{"/flag"}
	if got, want := settings.DataFileReadPaths(), []string{"/flag"}; !slices.Equal(got, want) {
		t.Errorf("DataFileReadPaths() with flag = %v, want %v", got, want)
	}

	settings.SettingsBase.Data.FileReadPaths = nil
	settings.FromFlags.Data.FileReadPaths = nil
	if got, want := settings.DataFileReadPaths(), []string{"/hard-default"}; !slices.Equal(got, want) {
		t.Errorf("DataFileReadPaths() with hard default = %v, want %v", got, want)
	}
}

func TestSettingsBase_MapEnvToStruct__fileReadPaths(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	test.IsolateEnv(t, map[string]string{
		"ARGUS_DATA_FILE_READ_PATHS": first + "," + second,
	})

	var settings SettingsBase
	if err := settings.MapEnvToStruct(); err != nil {
		t.Fatalf("SettingsBase.MapEnvToStruct() failed: %v", err)
	}
	if got, want := settings.Data.FileReadPaths, []string{first, second}; !slices.Equal(got, want) {
		t.Errorf("Data.FileReadPaths = %v, want %v", got, want)
	}
}
