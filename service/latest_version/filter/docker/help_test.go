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

package docker

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/internal/test"
	logtest "github.com/release-argus/Argus/internal/test/log"
)

func TestMain(m *testing.M) {
	logtest.InitLog()

	exitCode := m.Run()

	if len(logx.ExitCodeChannel()) > 0 {
		fmt.Println("exit code channel not empty")
		exitCode = 1
	}

	os.Exit(exitCode)
}

// plainDefaults returns plain defaults and hardDefaults for testing.
func plainDefaults() (*Defaults, *Defaults) {
	hardDefaults, _ := DecodeDefaults("yaml", nil, nil)
	hardDefaults.Default()
	defaults, _ := DecodeDefaults("yaml", nil, hardDefaults)

	defaults.SetDefaults(hardDefaults)

	return defaults, hardDefaults
}

// assertRegistriesNonNil fails t unless every registry is set.
func assertRegistriesNonNil(t *testing.T, registry RegistryDefaultsSet) {
	t.Helper()

	if testErr := test.AssertFields(
		t,
		[]test.FieldAssertion{
			{Name: "Registry.ECR", Got: registry.ECR, Mode: test.CompareNotNil},
			{Name: "Registry.GHCR", Got: registry.GHCR, Mode: test.CompareNotNil},
			{Name: "Registry.Hub", Got: registry.Hub, Mode: test.CompareNotNil},
			{Name: "Registry.Quay", Got: registry.Quay, Mode: test.CompareNotNil},
		},
		"Defaults",
	); testErr != nil {
		t.Fatal(testErr)
	}
}

// registryDefaultsString renders r as its type and address, or nil.
func registryDefaultsString(r RegistryDefaults) string {
	if r == nil {
		return "nil"
	}
	return fmt.Sprintf("%T(%p)", r, r)
}

type tokenData struct {
	token, queryToken string
	validUntil        time.Time
}

func getTokenData(auth RegistryAuth) (data tokenData) {
	switch a := auth.(type) {
	case *ECRAuth:
		data.queryToken = a.queryToken
		data.validUntil = a.validUntil
	case *GHCRAuth:
		data.token = a.GetToken()
		data.queryToken = a.queryToken
		data.validUntil = a.validUntil
	case *HubAuth:
		data.token = a.GetToken()
		data.queryToken = a.queryToken
		data.validUntil = a.validUntil
	case *QuayAuth:
		data.token = a.GetToken()
	}
	return
}

// assertInherited checks auth holds the token data of from when wantInherited, or had otherwise.
func assertInherited(
	t *testing.T,
	target string,
	had tokenData,
	auth, from RegistryAuth,
	wantInherited bool,
) {
	t.Helper()

	want := had
	if wantInherited {
		want = getTokenData(from)
	}
	got := getTokenData(auth)
	if got.token != want.token || got.queryToken != want.queryToken ||
		!got.validUntil.Equal(want.validUntil) {
		t.Errorf(
			"%s mismatch\n"+
				"got:  GetToken()=%q, queryToken=%q, validUntil=%s\n"+
				"want: GetToken()=%q, queryToken=%q, validUntil=%s",
			target,
			got.token, got.queryToken, got.validUntil,
			want.token, want.queryToken, want.validUntil,
		)
	}
}

// assertQueryTokenSet checks whether one layer of an auth defaults chain holds the given query token.
func assertQueryTokenSet(
	t *testing.T,
	target string,
	gotQueryToken string, gotValidUntil time.Time,
	queryToken string, validUntil time.Time,
	wantMatch bool,
) {
	t.Helper()

	not := ""
	if !wantMatch {
		not = "NOT "
	}
	if (gotQueryToken == queryToken) != wantMatch {
		t.Errorf(
			"%s.queryToken mismatch\ngot:  %q\nwant: %s%q",
			target, gotQueryToken, not, queryToken,
		)
	}
	if gotValidUntil.Equal(validUntil) != wantMatch {
		t.Errorf(
			"%s.validUntil mismatch\ngot:  %s\nwant: %s%s",
			target, gotValidUntil, not, validUntil,
		)
	}
}
