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

package shoutrrr

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
	metric "github.com/release-argus/Argus/web/metrics"
)

func TestShoutrrr_InitMetrics(t *testing.T) {
	// GIVEN a Shoutrrr
	tests := map[string]struct {
		serviceShoutrrr bool
		wantMetrics     bool
	}{
		"service Shoutrrr gives metrics": {
			serviceShoutrrr: true,
			wantMetrics:     true},
		"hardDefault/default/main Shoutrrr doesn't give metrics": {
			serviceShoutrrr: false,
			wantMetrics:     false},
		"service Shoutrrr with nil ServiceStatus doesn't give metrics": {
			serviceShoutrrr: false,
			wantMetrics:     false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			shoutrrr := testShoutrrr(false, true, false)
			*shoutrrr.ServiceStatus.ServiceID = name
			if !tc.serviceShoutrrr {
				shoutrrr.Main = nil
				shoutrrr.Defaults = nil
				shoutrrr.HardDefaults = nil
			}

			// WHEN the Prometheus metrics are initialised with initMetrics
			had := testutil.CollectAndCount(metric.NotifyMetric)
			shoutrrr.initMetrics()

			// THEN it can be collected
			// counters
			got := testutil.CollectAndCount(metric.NotifyMetric)
			want := 0
			if tc.wantMetrics {
				want = 2
			}
			if (got - had) != want {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(got - had), want)
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			shoutrrr := testShoutrrr(false, true, false)
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			shoutrrr := testShoutrrr(false, true, false)
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			shoutrrr := testShoutrrr(false, true, false)
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
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			shoutrrr := testShoutrrr(false, true, false)
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
		main            *Shoutrrr
		defaults        Shoutrrr
		hardDefaults    Shoutrrr
		serviceShoutrrr bool
		metricCount     int
		nilShoutrrr     bool
	}{
		"nil shoutrrr": {
			nilShoutrrr: true,
			metricCount: 0},
		"all lowercase keys": {
			id:              "lowercase",
			serviceShoutrrr: true,
			metricCount:     2,
			had: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"mixed-case keys": {
			id:              "mixed-case",
			serviceShoutrrr: true,
			metricCount:     2,
			had: map[string]string{
				"hello": "TEST123", "FOO": "bAr", "bIsh": "bash"},
			want: map[string]string{
				"hello": "TEST123", "foo": "bAr", "bish": "bash"}},
		"gives matching main": {
			id:              "matching-main",
			serviceShoutrrr: true,
			metricCount:     2,
			main:            &Shoutrrr{},
			giveMain:        true},
		"creates new main if none match": {
			id:              "no-matching-main",
			serviceShoutrrr: true,
			metricCount:     2,
			main:            nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			shoutrrr := testShoutrrr(false, true, false)
			shoutrrr.ID = tc.id
			serviceStatus := *shoutrrr.ServiceStatus
			*shoutrrr.ServiceStatus.ServiceID = name
			shoutrrr.Options = tc.had
			if tc.giveMain {
				tc.main.Options = tc.had
			}
			tc.defaults.URLFields = tc.had
			tc.hardDefaults.Params = tc.had
			if tc.nilShoutrrr {
				shoutrrr = nil
			}

			// WHEN Init is called on it
			hadC := testutil.CollectAndCount(metric.NotifyMetric)
			shoutrrr.Init(&serviceStatus, tc.main, &tc.defaults, &tc.hardDefaults)

			// THEN the Shoutrrr is initialised correctly
			// initMetrics - counters
			gotC := testutil.CollectAndCount(metric.NotifyMetric)
			if (gotC - hadC) != tc.metricCount {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(gotC - hadC), tc.metricCount)
			}
			if tc.nilShoutrrr {
				if shoutrrr != nil {
					t.Fatalf("nil shoutrrr should still be nil, not %v",
						shoutrrr)
				}
				return
			}
			// main
			if shoutrrr.Main != tc.main {
				if (tc.main == nil && shoutrrr.Main == nil) || tc.main != nil {
					t.Errorf("Main was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						tc.main, &shoutrrr.Main)
				}
			}
			// defaults
			if shoutrrr.Defaults != &tc.defaults {
				t.Errorf("Defaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&tc.defaults, shoutrrr.Defaults)
			}
			// hardDefaults
			if shoutrrr.HardDefaults != &tc.hardDefaults {
				t.Errorf("HardDefaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&tc.hardDefaults, shoutrrr.HardDefaults)
			}
			// status
			if shoutrrr.ServiceStatus != &serviceStatus {
				t.Errorf("Status was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
					&serviceStatus, shoutrrr.ServiceStatus)
			}
			for key := range tc.want {
				if tc.want[key] != shoutrrr.Options[key] {
					t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Options[key], tc.want, shoutrrr.Options)
				}
				if tc.want[key] != shoutrrr.Defaults.URLFields[key] {
					t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.Defaults.URLFields[key], tc.want, shoutrrr.Defaults.URLFields)
				}
				if tc.want[key] != shoutrrr.HardDefaults.Params[key] {
					t.Errorf("want: %q:%q\ngot:  %q:%q\n%v\n%v",
						key, tc.want[key], key, shoutrrr.HardDefaults.Params[key], tc.want, shoutrrr.HardDefaults.Params)
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
		mains        *Slice
		defaults     Slice
		hardDefaults Slice
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
				"fail": testShoutrrr(true, true, false),
				"pass": testShoutrrr(false, true, false)},
		},
		"slice with nil element and matching main": {
			slice: &Slice{
				"fail": nil},
			mains: &Slice{
				"fail": testShoutrrr(false, false, false)},
		},
		"have matching mains": {
			slice: &Slice{
				"fail": testShoutrrr(true, true, false),
				"pass": testShoutrrr(false, true, false)},
			mains: &Slice{
				"fail": testShoutrrr(false, true, false),
				"pass": testShoutrrr(true, true, false)},
		},
		"some matching mains": {
			slice: &Slice{
				"fail": testShoutrrr(true, true, false),
				"pass": testShoutrrr(false, true, false)},
			mains: &Slice{
				"other": testShoutrrr(false, true, false),
				"pass":  testShoutrrr(true, true, false)},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			log := util.NewJLog("WARN", false)
			if tc.slice != nil {
				for i := range *tc.slice {
					if (*tc.slice)[i] != nil {
						(*tc.slice)[i].ServiceStatus.ServiceID = stringPtr(name)
						(*tc.slice)[i].Options = tc.had
					}
				}
			}
			serviceStatus := svcstatus.Status{
				Fails:     svcstatus.Fails{Shoutrrr: map[string]*bool{}},
				ServiceID: stringPtr(name),
			}
			for i := range tc.defaults {
				tc.defaults[i].URLFields = tc.had
			}
			if tc.defaults == nil {
				tc.defaults = make(Slice)
			}
			for i := range tc.hardDefaults {
				tc.hardDefaults[i].Params = tc.had
			}
			if tc.hardDefaults == nil {
				tc.hardDefaults = make(Slice)
			}
			if tc.nilSlice {
				tc.slice = nil
			}

			// WHEN Init is called on it
			hadC := testutil.CollectAndCount(metric.NotifyMetric)
			tc.slice.Init(log, &serviceStatus, tc.mains, &tc.defaults, &tc.hardDefaults)

			// THEN the Shoutrrr is initialised correctly
			// initMetrics - counters
			gotC := testutil.CollectAndCount(metric.NotifyMetric)
			wantMetrics := 0
			if tc.slice != nil {
				wantMetrics = 2 * len(*tc.slice)
			}
			if (gotC - hadC) != wantMetrics {
				t.Errorf("%d Counter metrics's were initialised, expecting %d",
					(gotC - hadC), wantMetrics)
			}
			if tc.nilSlice {
				if tc.slice != nil {
					t.Fatalf("nil shoutrrr should still be nil, not %v",
						tc.slice)
				}
				return
			}

			for id := range *tc.slice {
				// main
				if (*tc.slice)[id].Main == nil || (tc.mains != nil && (*tc.mains)[id] != nil && (*tc.slice)[id].Main != (*tc.mains)[id]) {
					t.Errorf("Main was not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						(*tc.mains)[id], &(*tc.slice)[id].Main)
				}
				// defaults
				if tc.defaults[id] != nil && (*tc.slice)[id].Defaults != tc.defaults[id] {
					t.Errorf("Defaults were not handed to the Shoutrrr correctly\nwant: %v\ngot:  %v",
						tc.defaults[id], (*tc.slice)[id].Defaults)
				}
				// hardDefaults
				if tc.hardDefaults[id] != nil && (*tc.slice)[id].HardDefaults != tc.hardDefaults[id] {
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
