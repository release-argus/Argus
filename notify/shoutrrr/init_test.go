// Copyright [2025] [Argus]
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

package shoutrrr

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/status"
	serviceinfo "github.com/release-argus/Argus/service/status/info"
	"github.com/release-argus/Argus/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/web/metric"
)

func TestShoutrrrs_Metrics(t *testing.T) {
	// GIVEN Shoutrrrs.
	tests := map[string]struct {
		shoutrrrs *Shoutrrrs
	}{
		"nil": {
			shoutrrrs: nil},
		"empty": {
			shoutrrrs: &Shoutrrrs{}},
		"with one": {
			shoutrrrs: &Shoutrrrs{
				"foo": &Shoutrrr{}}},
		"multiple": {
			shoutrrrs: &Shoutrrrs{
				"bish": &Shoutrrr{},
				"bash": &Shoutrrr{},
				"bosh": &Shoutrrr{}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - Cannot run in parallel since we're testing metrics.

			if tc.shoutrrrs != nil {
				for name, s := range *tc.shoutrrrs {
					s.ID = name
					s.ServiceStatus = &status.Status{
						ServiceInfo: serviceinfo.ServiceInfo{
							ID: name + "-service"}}
					s.Main = &Defaults{}
					s.Type = "gotify"
				}
			}

			// WHEN the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			tc.shoutrrrs.InitMetrics()

			// THEN it can be counted.
			got := testutil.CollectAndCount(metric.NotifyResultTotal)
			want := had
			if tc.shoutrrrs != nil {
				want += 2 * len(*tc.shoutrrrs)
			}
			if got != want {
				t.Errorf("%s\nInitMetrics() mismatch\nwant: %d counter metrics\ngot:  %d",
					packageName, want, got)
			}

			// AND the metrics can be deleted.
			tc.shoutrrrs.DeleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyResultTotal)
			if got != had {
				t.Errorf("%s\nDeleteMetrics() mismatch\nwant: %d counter metrics\ngot:  %d",
					packageName, want, got)
			}
		})
	}
}

func TestShoutrrr_Metrics(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := []string{
		"a service",
		"another service",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name

			// WHEN the Prometheus metrics are initialised with initMetrics.
			had := testutil.CollectAndCount(metric.NotifyResultTotal)
			shoutrrr.initMetrics()

			// THEN it can be collected.
			// counters:
			got := testutil.CollectAndCount(metric.NotifyResultTotal)
			want := 2
			if (got - had) != want {
				t.Errorf("%s\nInitMetrics() mismatch\nwant: %d counter metrics\ngot:  %d",
					packageName, want, got-had)
			}

			// AND it can be deleted.
			shoutrrr.deleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyResultTotal)
			if got != had {
				t.Errorf("%s\ndeleteMetrics() mismatch\nwant: %d counter metrics\ngot:  %d",
					packageName, got, had)
			}
		})
	}
}

func TestShoutrrr_InitOptions(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		had, want map[string]string
	}{
		"all lowercase keys": {
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name
			shoutrrr.Options = tc.had

			// WHEN initOptions is called on it.
			shoutrrr.initOptions()

			// THEN the keys in the map will have been converted to lowercase.
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.Options)) {
				t.Fatalf("%s\nwant: %v\ngot:  %v",
					packageName, tc.want, shoutrrr.Options)
			}
		})
	}
}

func TestShoutrrr_InitURLFields(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		had, want map[string]string
	}{
		"all lowercase keys": {
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name
			shoutrrr.URLFields = tc.had

			// WHEN initURLFields is called on it.
			shoutrrr.initURLFields()

			// THEN the keys in the map will have been converted to lowercase.
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.URLFields)) {
				t.Fatalf("%s\nwant: %v\ngot:  %v",
					packageName, tc.want, shoutrrr.URLFields)
			}
		})
	}
}

func TestShoutrrr_InitParams(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		had, want map[string]string
	}{
		"all lowercase keys": {
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name
			shoutrrr.Params = tc.had

			// WHEN initParams is called on it.
			shoutrrr.initParams()

			// THEN the keys in the map will have been converted to lowercase.
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.Params)) {
				t.Fatalf("%s\nwant: %v\ngot:  %v",
					packageName, tc.want, shoutrrr.Params)
			}
		})
	}
}

func TestShoutrrr_InitMaps(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		had, want   map[string]string
		nilShoutrrr bool
	}{
		"nil shoutrrr": {
			nilShoutrrr: true},
		"all lowercase keys": {
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ServiceStatus.ServiceInfo.ID = name
			shoutrrr.Options = tc.had
			shoutrrr.URLFields = tc.had
			shoutrrr.Params = tc.had
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN InitMaps is called on it.
			shoutrrr.InitMaps()

			// THEN the keys in the options/urlFields/params maps will have been converted to lowercase.
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf("%s\nnil shoutrrr should remain nil\ngot: %v",
						packageName, shoutrrr)
				}
				return
			}
			errStr := "%s\nmismatch on %s\nwant: %v\ngot:  %v"
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.Options)) {
				t.Fatalf(errStr,
					packageName, "Options", tc.want, shoutrrr.Options)
			}
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.URLFields)) {
				t.Fatalf(errStr,
					packageName, "URLFields", tc.want, shoutrrr.URLFields)
			}
			if !test.EqualSlices(util.SortedKeys(tc.want), util.SortedKeys(shoutrrr.Params)) {
				t.Fatalf(errStr,
					packageName, "Params", tc.want, shoutrrr.Params)
			}
			errStr = "%s\nmismatch on %s\nwant: %q:%q\ngot:  %q:%q\n%v\n%v"
			for key := range tc.want {
				if shoutrrr.Options[key] != tc.want[key] {
					t.Fatalf(errStr,
						packageName, "Options", key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
				if shoutrrr.URLFields[key] != tc.want[key] {
					t.Fatalf(errStr,
						packageName, "URLFields", key, tc.want[key], key, shoutrrr.URLFields[key], tc.want, shoutrrr.URLFields)
				}
				if shoutrrr.Params[key] != tc.want[key] {
					t.Fatalf(errStr,
						packageName, "Params", key, tc.want[key], key, shoutrrr.Params[key], tc.want, shoutrrr.Params)
				}
			}
		})
	}
}

func TestShoutrrr_Init(t *testing.T) {
	// GIVEN a Shoutrrr.
	tests := map[string]struct {
		id                           string
		had, want                    map[string]string
		giveMain                     bool
		main                         *Defaults
		serviceShoutrrr, nilShoutrrr bool
	}{
		"nil shoutrrr": {
			nilShoutrrr: true},
		"all lowercase keys": {
			id:              "lowercase",
			serviceShoutrrr: true,
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			id:              "mixed-case",
			serviceShoutrrr: true,
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"gives matching main": {
			id:              "matching-main",
			serviceShoutrrr: true,
			main:            &Defaults{},
			giveMain:        true},
		"creates new main if none match": {
			id:              "no-matching-main",
			serviceShoutrrr: true,
			main:            nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			shoutrrr := testShoutrrr(false, false)
			shoutrrr.ID = tc.id
			serviceStatus := shoutrrr.ServiceStatus
			shoutrrr.ServiceStatus.ServiceInfo.ID = name
			if tc.giveMain {
				tc.main.Options = tc.had
			}
			shoutrrr.Options = map[string]string{}
			shoutrrr.Params = map[string]string{}
			shoutrrr.URLFields = map[string]string{}
			defaults := NewDefaults(
				"",
				make(map[string]string), make(map[string]string), make(map[string]string))
			hardDefaults := NewDefaults(
				"",
				make(map[string]string), make(map[string]string), make(map[string]string))
			for key := range tc.had {
				shoutrrr.Options[key] = tc.had[key]
				defaults.Options[key] = tc.had[key]
				hardDefaults.Options[key] = tc.had[key]
				shoutrrr.Params[key] = tc.had[key]
				defaults.Params[key] = tc.had[key]
				hardDefaults.Params[key] = tc.had[key]
				shoutrrr.URLFields[key] = tc.had[key]
				defaults.URLFields[key] = tc.had[key]
				hardDefaults.URLFields[key] = tc.had[key]
			}
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN Init is called on it.
			shoutrrr.Init(
				serviceStatus,
				tc.main,
				defaults, hardDefaults)

			// THEN the Shoutrrr is initialised correctly:
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf("%s\nnil shoutrrr should remain nil\ngot: %v",
						packageName, shoutrrr)
				}
				return
			}
			errStr := "%s\n%s not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v"
			// 	main:
			if shoutrrr.Main != tc.main && tc.giveMain {
				t.Errorf(errStr,
					packageName, "Main", tc.main, &shoutrrr.Main)
			}
			// 	defaults:
			if shoutrrr.Defaults != defaults {
				t.Errorf(errStr,
					packageName, "Defaults", &defaults, shoutrrr.Defaults)
			}
			// 	hardDefaults:
			if shoutrrr.HardDefaults != hardDefaults {
				t.Errorf(errStr,
					packageName, "HardDefaults", &hardDefaults, shoutrrr.HardDefaults)
			}
			// 	status:
			if shoutrrr.ServiceStatus != serviceStatus {
				t.Errorf(errStr,
					packageName, "Status", &serviceStatus, shoutrrr.ServiceStatus)
			}
			errStr = "%s\nmismatch on %s\nwant: %q:%q\ngot:  %q:%q\n%v\n%v"
			for key := range tc.want {
				if shoutrrr.Options[key] != tc.want[key] {
					t.Errorf(errStr,
						packageName, "Options", key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
				if shoutrrr.URLFields[key] != tc.want[key] {
					t.Errorf(errStr,
						packageName, "URLFields", key, tc.want[key], key, shoutrrr.URLFields[key], tc.want, shoutrrr.URLFields)
				}
				if shoutrrr.Params[key] != tc.want[key] {
					t.Errorf(errStr,
						packageName, "Params", key, tc.want[key], key, shoutrrr.Params[key], tc.want, shoutrrr.Params)
				}
			}
		})
	}
}

func TestShoutrrrs_Init(t *testing.T) {
	// GIVEN Shoutrrrs.
	tests := map[string]struct {
		nilMap                 bool
		shoutrrrs              *Shoutrrrs
		had, want              map[string]string
		mains                  *ShoutrrrsDefaults
		defaults, hardDefaults ShoutrrrsDefaults
	}{
		"nil slice": {
			shoutrrrs: nil,
			nilMap:    true,
		},
		"empty slice": {
			shoutrrrs: &Shoutrrrs{},
		},
		"nil mains": {
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
		},
		"slice with nil element and matching main": {
			shoutrrrs: &Shoutrrrs{
				"fail": nil},
			mains: &ShoutrrrsDefaults{
				"fail": testDefaults(false, false)},
		},
		"have matching mains": {
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
			mains: &ShoutrrrsDefaults{
				"fail": testDefaults(false, false),
				"pass": testDefaults(true, false)},
		},
		"some matching mains": {
			shoutrrrs: &Shoutrrrs{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
			mains: &ShoutrrrsDefaults{
				"other": testDefaults(false, false),
				"pass":  testDefaults(true, false)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			serviceStatus := status.Status{}
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				mainCount, 0, 0,
				name, "", "",
				&dashboard.Options{})
			for i := range tc.defaults {
				tc.defaults[i].URLFields = tc.had
			}
			if tc.defaults == nil {
				tc.defaults = make(ShoutrrrsDefaults)
			}
			for i := range tc.hardDefaults {
				tc.hardDefaults[i].Params = tc.had
			}
			if tc.hardDefaults == nil {
				tc.hardDefaults = make(ShoutrrrsDefaults)
			}
			if tc.nilMap {
				tc.shoutrrrs = nil
			}

			// WHEN Init is called on it.
			tc.shoutrrrs.Init(
				&serviceStatus,
				tc.mains, &tc.defaults, &tc.hardDefaults)

			// THEN the Shoutrrr is initialised correctly:
			if tc.nilMap {
				if tc.shoutrrrs != nil {
					t.Fatalf("%s\nnil shoutrrr should still be nil\ngot: %v",
						packageName, tc.shoutrrrs)
				}
				return
			}

			for id := range *tc.shoutrrrs {
				errStr := "%s\n%s not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v"
				// 	main:
				if (*tc.shoutrrrs)[id].Main == nil ||
					(tc.mains != nil && (*tc.mains)[id] != nil && (*tc.shoutrrrs)[id].Main != (*tc.mains)[id]) {
					t.Errorf(errStr,
						packageName, "Main", (*tc.mains)[id], &(*tc.shoutrrrs)[id].Main)
				}
				// 	defaults:
				if tc.defaults[id] != nil &&
					(*tc.shoutrrrs)[id].Defaults != tc.defaults[id] {
					t.Errorf(errStr,
						packageName, "Defaults", tc.defaults[id], (*tc.shoutrrrs)[id].Defaults)
				}
				// 	hardDefaults:
				if tc.hardDefaults[id] != nil &&
					(*tc.shoutrrrs)[id].HardDefaults != tc.hardDefaults[id] {
					t.Errorf(errStr,
						packageName, "HardDefaults", tc.hardDefaults[id], (*tc.shoutrrrs)[id].HardDefaults)
				}
				// 	status:
				if (*tc.shoutrrrs)[id].ServiceStatus != &serviceStatus {
					t.Errorf(errStr,
						packageName, "Status", &serviceStatus, (*tc.shoutrrrs)[id].ServiceStatus)
				}
				if &(*tc.shoutrrrs)[id].ServiceStatus.Fails.Shoutrrr != (*tc.shoutrrrs)[id].Failed {
					t.Errorf(errStr,
						packageName, "Status.Fails", &(*tc.shoutrrrs)[id].ServiceStatus.Fails.Shoutrrr, (*tc.shoutrrrs)[id].Failed)
				}
				errStr = "%s\nmismatch on %s\nwant: %q:%q\ngot:  %q:%q\n%v\n%v"
				for key := range tc.want {
					if (*tc.shoutrrrs)[id].Options[key] != tc.want[key] {
						t.Errorf(errStr,
							packageName, "Options",
							key, tc.want[key], key, (*tc.shoutrrrs)[id].Options[key], tc.want, (*tc.shoutrrrs)[id].Options)
					}
					if (*tc.shoutrrrs)[id].Defaults.URLFields[key] != tc.want[key] {
						t.Errorf(errStr,
							packageName, "URLFields",
							key, tc.want[key], key, (*tc.shoutrrrs)[id].Defaults.URLFields[key], tc.want, (*tc.shoutrrrs)[id].Defaults.URLFields)
					}
					if (*tc.shoutrrrs)[id].HardDefaults.Params[key] != tc.want[key] {
						t.Errorf(errStr,
							packageName, "Params",
							key, tc.want[key], key, (*tc.shoutrrrs)[id].HardDefaults.Params[key], tc.want, (*tc.shoutrrrs)[id].HardDefaults.Params)
					}
				}
			}
		})
	}
}
