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

//go:build integration

package shoutrrr

import (
	"fmt"
	"strings"
	"testing"

	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

func TestShoutrrr_Send(t *testing.T) {
	// GIVEN a Shoutrrr to try and send
	testLogging("DEBUG")
	tests := map[string]struct {
		shoutrrr  *Shoutrrr
		svcStatus *svcstatus.Status
		delay     string
		useDelay  bool
		message   string
		retries   int
		deleting  bool
		errRegex  string
	}{
		"empty": {
			shoutrrr: &Shoutrrr{},
			errRegex: "failed to create Shoutrrr sender",
		},
		"valid, empty message": {
			shoutrrr: testShoutrrr(false, false, false),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			errRegex: "",
		},
		"valid, with message": {
			shoutrrr: testShoutrrr(false, false, false),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			message:  "__name__",
			errRegex: "",
		},
		"valid, with message, with delay": {
			shoutrrr: testShoutrrr(false, false, false),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			message:  "__name__",
			useDelay: true,
			delay:    "1s",
			errRegex: "",
		},
		"invalid https cert": {
			shoutrrr: testShoutrrr(false, false, true),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			errRegex: "x509",
		},
		"failing": {
			shoutrrr: testShoutrrr(true, false, true),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			retries:  1,
			errRegex: "invalid gotify token .* x 2",
		},
		"deleting": {
			shoutrrr: testShoutrrr(true, false, true),
			svcStatus: &svcstatus.Status{
				ServiceID: stringPtr("Testing")},
			deleting: true,
			errRegex: "",
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		tc.shoutrrr.Init(
			tc.svcStatus,
			&Shoutrrr{}, &Shoutrrr{}, &Shoutrrr{})

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if tc.shoutrrr.Failed == nil &&
				tc.shoutrrr.ServiceStatus != nil {
				tc.shoutrrr.Failed = &map[string]*bool{
					*tc.shoutrrr.ServiceStatus.ServiceID: boolPtr(false)}
			}
			if tc.shoutrrr.ServiceStatus != nil {
				tc.shoutrrr.ServiceStatus.Deleting = tc.deleting
			}
			if tc.shoutrrr.Options == nil {
				tc.shoutrrr.Options = map[string]string{}
			}
			if tc.delay != "" {
				tc.shoutrrr.Options["delay"] = tc.delay
			}
			tc.shoutrrr.Options["max_tries"] = fmt.Sprint(tc.retries + 1)

			// WHEN send attempted
			msg := strings.ReplaceAll(tc.message, "__name__", name)
			err := tc.shoutrrr.Send(
				"test",
				msg,
				&util.ServiceInfo{ID: "Testing"},
				tc.useDelay)

			// THEN any error should match the expected regex
			if err == nil {
				if tc.errRegex != "" {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if !util.RegexCheck(tc.errRegex, err.Error()) {
				t.Errorf("invalid error:\nwant: %s\ngot:  %s",
					tc.errRegex, util.ErrorToString(err))
			}
		})
	}
}
