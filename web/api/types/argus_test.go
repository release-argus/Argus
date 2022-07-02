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

package types

import (
	"testing"

	"github.com/release-argus/Argus/utils"
)

func TestNotifyCensorWithNil(t *testing.T) {
	// GIVEN a nil Notify
	var notify *Notify

	// WHEN Censor is called on it
	got := notify.Censor()

	// THEN nil is returned
	if got != nil {
		t.Errorf("Censor on %v should've return nil, not %v",
			notify, got)
	}
}

func TestNotifyCensorWithNonCensorableURLFieldsAndParams(t *testing.T) {
	// GIVEN a Notify with no URLFields or Params that should be censored
	notify := Notify{
		URLFields: map[string]string{
			"username": "a",
			"port":     "bb",
			"host":     "ccc",
		},
		Params: map[string]string{
			"botname": "a",
			"color":   "bb",
		},
	}

	// WHEN Censor is called on it
	got := notify.Censor()

	// THEN none of the URLFields or Params are censored
	wantCensor := "<secret>"
	for key, value := range (*got).URLFields {
		if value == wantCensor {
			t.Errorf("%q should not have been censored to %q",
				key, value)
		}
	}
	for key, value := range (*got).Params {
		if value != "" &&
			value == wantCensor {
			t.Errorf("%q should not have been censored to %q",
				key, value)
		}
	}
}

func TestNotifyCensorURLFields(t *testing.T) {
	// GIVEN a Notify with URL Fields to censor
	notify := testNotify()
	should_be_censored := []string{
		"apikey",
		"botkey",
		"password",
		"token",
		"tokena",
		"tokenb",
		"webhookid",
	}

	// WHEN Censor is called on it
	got := notify.Censor()

	// THEN all the confidential URL Fields are censored
	if got == nil {
		t.Errorf("Censor on %v shouldn't have returned a censored %v",
			notify, got)
	}
	wantCensor := "<secret>"
	for key, value := range (*got).URLFields {
		if value != "" &&
			utils.Contains(should_be_censored, key) &&
			value != wantCensor {
			t.Errorf("%q should have been censored to %q, not %q",
				key, wantCensor, value)
		}
	}
}

func TestNotifyCensorParams(t *testing.T) {
	// GIVEN a Notify with Params to censor
	notify := testNotify()
	should_be_censored := []string{
		"devices",
		"host",
	}

	// WHEN Censor is called on it
	got := notify.Censor()

	// THEN all the confidential Params are censored
	if got == nil {
		t.Errorf("Censor on %v shouldn't have returned a censored %v",
			notify, got)
	}
	wantCensor := "<secret>"
	for key, value := range (*got).Params {
		if value != "" &&
			utils.Contains(should_be_censored, key) &&
			value != wantCensor {
			t.Errorf("%q should have been censored to %q, not %q",
				key, wantCensor, value)
		}
	}
}

func TestNotifySliceCensorWithNil(t *testing.T) {
	// GIVEN a nil Notify Slice
	var slice *NotifySlice

	// WHEN Censor is called on it
	got := slice.Censor()

	// THEN nil is returned
	if got != nil {
		t.Errorf("Censor on %v should've returned nil, not %v",
			slice, got)
	}
}

func TestNotifySliceCensorWithNonNil(t *testing.T) {
	// GIVEN a non-nil Notify Slice
	notify := testNotify()
	slice := NotifySlice{
		"a": &notify,
		"b": &notify,
		"c": &notify,
	}

	// WHEN Censor is called on it
	got := slice.Censor()

	// THEN len(slice) is returned, censored
	if len(*got) != len(slice) {
		t.Errorf("Censor should've returned %d censored elements, not %d",
			len(slice), len(*got))
	}
}
