// Copyright [2023] [Argus]
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
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/test"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestSlice_Metrics(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		slice *Slice
	}{
		"nil": {
			slice: nil},
		"empty": {
			slice: &Slice{}},
		"with one": {
			slice: &Slice{
				"foo": &Shoutrrr{}}},
		"multiple": {
			slice: &Slice{
				"bish": &Shoutrrr{},
				"bash": &Shoutrrr{},
				"bosh": &Shoutrrr{}}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// t.Parallel() - not parallel as we are testing metrics

			if tc.slice != nil {
				for name, s := range *tc.slice {
					s.ID = name
					s.ServiceStatus = &svcstatus.Status{ServiceID: test.StringPtr(name + "-service")}
					s.Main = &ShoutrrrDefaults{}
					s.Type = "gotify"
				}
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			had := testutil.CollectAndCount(metric.NotifyMetric)
			tc.slice.InitMetrics()

			// THEN it can be counted
			got := testutil.CollectAndCount(metric.NotifyMetric)
			want := had
			if tc.slice != nil {
				want += 2 * len(*tc.slice)
			}
			if got != want {
				t.Errorf("got %d metrics, expecting %d",
					got, want)
			}

			// AND the metrics can be deleted
			tc.slice.DeleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyMetric)
			if got != had {
				t.Errorf("deleted metrics but got %d, expecting %d",
					got, want)
			}
		})
	}
}

func TestShoutrrr_Metrics(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := []string{
		"a service",
		"another service",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {

			shoutrrr := testShoutrrr(false, false)
			*shoutrrr.ServiceStatus.ServiceID = name

			// WHEN the Prometheus metrics are initialised with initMetrics
			had := testutil.CollectAndCount(metric.NotifyMetric)
			shoutrrr.initMetrics()

			// THEN it can be collected
			// counters
			got := testutil.CollectAndCount(metric.NotifyMetric)
			want := 2
			if (got - had) != want {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(got - had), want)
			}

			// AND it can be deleted
			shoutrrr.deleteMetrics()
			got = testutil.CollectAndCount(metric.NotifyMetric)
			if got != had {
				t.Errorf("Counter metrics's were deleted, got %d. expecting %d",
					got, had)
			}
		})
	}
}

func TestShoutrrr_InitOptions(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		had  map[string]string
		want map[string]string
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
			*shoutrrr.ServiceStatus.ServiceID = name
			shoutrrr.Options = tc.had

			// WHEN initOptions is called on it
			shoutrrr.initOptions()

			// THEN the keys in the map will have been converted to lowercase
			if len(tc.want) != len(shoutrrr.Options) {
				t.Fatalf("want: %v\ngot: %v",
					tc.want, shoutrrr.Options)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.Options[key] {
					t.Fatalf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
			}
		})
	}
}

func TestShoutrrr_InitURLFields(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		had  map[string]string
		want map[string]string
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
			*shoutrrr.ServiceStatus.ServiceID = name
			shoutrrr.URLFields = tc.had

			// WHEN initURLFields is called on it
			shoutrrr.initURLFields()

			// THEN the keys in the map will have been converted to lowercase
			if len(tc.want) != len(shoutrrr.URLFields) {
				t.Fatalf("want: %v\ngot: %v",
					tc.want, shoutrrr.URLFields)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.URLFields[key] {
					t.Fatalf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.URLFields[key], tc.want, shoutrrr.URLFields)
				}
			}
		})
	}
}

func TestShoutrrr_InitParams(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		had  map[string]string
		want map[string]string
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
			*shoutrrr.ServiceStatus.ServiceID = name
			shoutrrr.Params = tc.had

			// WHEN initParams is called on it
			shoutrrr.initParams()

			// THEN the keys in the map will have been converted to lowercase
			if len(tc.want) != len(shoutrrr.Params) {
				t.Fatalf("want: %v\ngot: %v",
					tc.want, shoutrrr.Params)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.Params[key] {
					t.Fatalf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Params[key], tc.want, shoutrrr.Params)
				}
			}
		})
	}
}

func TestShoutrrr_InitMaps(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		had         map[string]string
		want        map[string]string
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
			*shoutrrr.ServiceStatus.ServiceID = name
			shoutrrr.Options = tc.had
			shoutrrr.URLFields = tc.had
			shoutrrr.Params = tc.had
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN InitMaps is called on it
			shoutrrr.InitMaps()

			// THEN the keys in the options/urlfields/params maps will have been converted to lowercase
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf("nil shoutrrr should still be nil, not %v",
						shoutrrr)
				}
				return
			}
			if len(tc.want) != len(shoutrrr.Options) {
				t.Fatalf("Options:\nwant: %v\ngot: %v",
					tc.want, shoutrrr.Options)
			}
			if len(tc.want) != len(shoutrrr.URLFields) {
				t.Fatalf("URLFields:\nwant: %v\ngot: %v",
					tc.want, shoutrrr.URLFields)
			}
			if len(tc.want) != len(shoutrrr.Params) {
				t.Fatalf("Params:\nwant: %v\ngot: %v",
					tc.want, shoutrrr.Params)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.Options[key] {
					t.Fatalf("Options:\nwant: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
				if tc.want[key] != shoutrrr.URLFields[key] {
					t.Fatalf("URLFields:\nwant: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.URLFields[key], tc.want, shoutrrr.URLFields)
				}
				if tc.want[key] != shoutrrr.Params[key] {
					t.Fatalf("Params:\nwant: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Params[key], tc.want, shoutrrr.Params)
				}
			}
		})
	}
}

func TestShoutrrr_Init(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		id              string
		had             map[string]string
		want            map[string]string
		giveMain        bool
		main            *ShoutrrrDefaults
		serviceShoutrrr bool
		nilShoutrrr     bool
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
			main:            &ShoutrrrDefaults{},
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
			*shoutrrr.ServiceStatus.ServiceID = name
			if tc.giveMain {
				tc.main.Options = tc.had
			}
			shoutrrr.Options = map[string]string{}
			shoutrrr.Params = map[string]string{}
			shoutrrr.URLFields = map[string]string{}
			defaults := NewDefaults(
				"", &map[string]string{}, &map[string]string{}, &map[string]string{})
			hardDefaults := NewDefaults(
				"", &map[string]string{}, &map[string]string{}, &map[string]string{})
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

			// WHEN Init is called on it
			shoutrrr.Init(
				serviceStatus,
				tc.main, defaults, hardDefaults)

			// THEN the Shoutrrr is initialised correctly
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf("nil shoutrrr should still be nil, not %v",
						shoutrrr)
				}
				return
			}
			// main
			if shoutrrr.Main != tc.main && tc.giveMain {
				t.Errorf("Main was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					tc.main, &shoutrrr.Main)
			}
			// defaults
			if shoutrrr.Defaults != defaults {
				t.Errorf("Defaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&defaults, shoutrrr.Defaults)
			}
			// hardDefaults
			if shoutrrr.HardDefaults != hardDefaults {
				t.Errorf("HardDefaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&hardDefaults, shoutrrr.HardDefaults)
			}
			// status
			if shoutrrr.ServiceStatus != serviceStatus {
				t.Errorf("Status was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&serviceStatus, shoutrrr.ServiceStatus)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.Options[key] {
					t.Errorf("want: %q:%q, got:  %q:%q\nwant: %v\ngot: %v",
						key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
				if tc.want[key] != shoutrrr.URLFields[key] {
					t.Errorf("want: %q:%q, got:  %q:%q\nwant: %v\ngot: %v",
						key, tc.want[key], key, shoutrrr.URLFields[key], tc.want, shoutrrr.URLFields)
				}
				if tc.want[key] != shoutrrr.Params[key] {
					t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Params[key], tc.want, shoutrrr.Params)
				}
			}
		})
	}
}

func TestSlice_Init(t *testing.T) {
	// GIVEN a Slice
	tests := map[string]struct {
		nilSlice     bool
		slice        *Slice
		had          map[string]string
		want         map[string]string
		mains        *SliceDefaults
		defaults     SliceDefaults
		hardDefaults SliceDefaults
	}{
		"nil slice": {
			slice:    nil,
			nilSlice: true,
		},
		"empty slice": {
			slice: &Slice{},
		},
		"nil mains": {
			slice: &Slice{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
		},
		"slice with nil element and matching main": {
			slice: &Slice{
				"fail": nil},
			mains: &SliceDefaults{
				"fail": testShoutrrrDefaults(false, false)},
		},
		"have matching mains": {
			slice: &Slice{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
			mains: &SliceDefaults{
				"fail": testShoutrrrDefaults(false, false),
				"pass": testShoutrrrDefaults(true, false)},
		},
		"some matching mains": {
			slice: &Slice{
				"fail": testShoutrrr(true, false),
				"pass": testShoutrrr(false, false)},
			mains: &SliceDefaults{
				"other": testShoutrrrDefaults(false, false),
				"pass":  testShoutrrrDefaults(true, false)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.slice != nil {
				for i := range *tc.slice {
					if (*tc.slice)[i] != nil {
						(*tc.slice)[i].ServiceStatus.ServiceID = test.StringPtr(name)
						(*tc.slice)[i].Options = tc.had
					}
				}
			}
			serviceStatus := svcstatus.Status{}
			mainCount := 0
			if tc.mains != nil {
				mainCount = len(*tc.mains)
			}
			serviceStatus.Init(
				mainCount, 0, 0,
				&name,
				nil)
			for i := range tc.defaults {
				tc.defaults[i].URLFields = tc.had
			}
			if tc.defaults == nil {
				tc.defaults = make(SliceDefaults)
			}
			for i := range tc.hardDefaults {
				tc.hardDefaults[i].Params = tc.had
			}
			if tc.hardDefaults == nil {
				tc.hardDefaults = make(SliceDefaults)
			}
			if tc.nilSlice {
				tc.slice = nil
			}

			// WHEN Init is called on it
			tc.slice.Init(
				&serviceStatus,
				tc.mains, &tc.defaults, &tc.hardDefaults)

			// THEN the Shoutrrr is initialised correctly
			if tc.nilSlice {
				if tc.slice != nil {
					t.Fatalf("nil shoutrrr should still be nil, not %v",
						tc.slice)
				}
				return
			}

			for id := range *tc.slice {
				// main
				if (*tc.slice)[id].Main == nil ||
					(tc.mains != nil && (*tc.mains)[id] != nil && (*tc.slice)[id].Main != (*tc.mains)[id]) {
					t.Errorf("Main was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						(*tc.mains)[id], &(*tc.slice)[id].Main)
				}
				// defaults
				if tc.defaults[id] != nil &&
					(*tc.slice)[id].Defaults != tc.defaults[id] {
					t.Errorf("Defaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						tc.defaults[id], (*tc.slice)[id].Defaults)
				}
				// hardDefaults
				if tc.hardDefaults[id] != nil &&
					(*tc.slice)[id].HardDefaults != tc.hardDefaults[id] {
					t.Errorf("HardDefaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						tc.hardDefaults[id], (*tc.slice)[id].HardDefaults)
				}
				// status
				if (*tc.slice)[id].ServiceStatus != &serviceStatus {
					t.Errorf("Status was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						&serviceStatus, (*tc.slice)[id].ServiceStatus)
				}
				if &(*tc.slice)[id].ServiceStatus.Fails.Shoutrrr != (*tc.slice)[id].Failed {
					t.Errorf("Status was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						&(*tc.slice)[id].ServiceStatus.Fails.Shoutrrr, (*tc.slice)[id].Failed)
				}
				for key := range tc.want {
					if tc.want[key] != (*tc.slice)[id].Options[key] {
						t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
							key, tc.want[key], key, (*tc.slice)[id].Options[key], tc.want, (*tc.slice)[id].Options)
					}
					if tc.want[key] != (*tc.slice)[id].Defaults.URLFields[key] {
						t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
							key, tc.want[key], key, (*tc.slice)[id].Defaults.URLFields[key], tc.want, (*tc.slice)[id].Defaults.URLFields)
					}
					if tc.want[key] != (*tc.slice)[id].HardDefaults.Params[key] {
						t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
							key, tc.want[key], key, (*tc.slice)[id].HardDefaults.Params[key], tc.want, (*tc.slice)[id].HardDefaults.Params)
					}
				}
			}
		})
	}
}
