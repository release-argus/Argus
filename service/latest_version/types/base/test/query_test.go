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
	"errors"
	"strings"
	"testing"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/service/dashboard"
	"github.com/release-argus/Argus/service/latest_version/types/base"
	"github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// fakeLookup is a [base.Interface] whose Query finds version, fails with err, or announces.
type fakeLookup struct {
	base.Interface
	status   *status.Status
	version  string
	err      error
	announce bool
}

func (l *fakeLookup) GetStatus() *status.Status { return l.status }

func (l *fakeLookup) GetType() string { return "fake" }

func (l *fakeLookup) Query(_ bool, _ logx.LogFrom) (bool, error) {
	if l.err != nil {
		return false, l.err
	}
	if l.announce {
		l.status.AnnounceQuery()
	}
	l.status.SetLatestVersion(l.version, "", false)
	return true, nil
}

func TestAssertQueryNoSideEffects(t *testing.T) {
	tests := []struct {
		name       string
		lookup     *fakeLookup
		fatalRegex string
		errorRegex string
	}{
		{
			name:       "no side effects",
			lookup:     &fakeLookup{version: "1.2.3"},
			fatalRegex: `^$`,
			errorRegex: `^$`,
		},
		{
			name:       "Query fails",
			lookup:     &fakeLookup{err: errors.New("no version found")},
			fatalRegex: `^Lookup\.Query\(\) failed: no version found$`,
			errorRegex: `^$`,
		},
		{
			name:       "announces",
			lookup:     &fakeLookup{version: "1.2.3", announce: true},
			fatalRegex: `^$`,
			errorRegex: `^len\(Status\.AnnounceChannel\) mismatch`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.lookup.status = status.New(
				nil, nil, nil,
				"",
				"", "",
				"", "",
				"",
				&dashboard.Options{},
			)

			fake := &test.FakeT{TestName: t.Name(), Abort: true}
			t.Cleanup(fake.RunCleanups)

			fake.Aborted(func() { AssertQueryNoSideEffects(fake, tc.lookup, "1.2.2", "1.2.3") })

			if got := strings.Join(fake.Fatals, "\n"); !util.RegexCheck(tc.fatalRegex, got) {
				t.Errorf("FakeT.Fatals mismatch\ngot:  %q\nwant: %q", got, tc.fatalRegex)
			}
			if got := strings.Join(fake.Errors, "\n"); !util.RegexCheck(tc.errorRegex, got) {
				t.Errorf("FakeT.Errors mismatch\ngot:  %q\nwant: %q", got, tc.errorRegex)
			}
		})
	}
}
