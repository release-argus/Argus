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

//go:build integration

package httpx

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/release-argus/Argus/internal/test"
	"github.com/release-argus/Argus/util"
	"github.com/release-argus/Argus/util/errfmt"
)

func TestClient_Get__certValidation(t *testing.T) {
	tests := []struct {
		name        string
		client      *http.Client
		invalidCert bool
		errRegex    string
	}{
		{
			name:     "Client accepts valid certificate",
			client:   Client,
			errRegex: `^$`,
		},
		{
			name:        "Client rejects invalid certificate",
			client:      Client,
			invalidCert: true,
			errRegex:    `x509`,
		},
		{
			name:     "InsecureClient accepts valid certificate",
			client:   InsecureClient,
			errRegex: `^$`,
		},
		{
			name:        "InsecureClient accepts invalid certificate",
			client:      InsecureClient,
			invalidCert: true,
			errRegex:    `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url := test.LookupPlain.URLValid
			if tc.invalidCert {
				url = test.LookupPlain.URLInvalid
			}

			for try := 0; ; try++ {
				err := getPlain(tc.client, url)
				e := errfmt.FormatError(err)
				if strings.Contains(e, "context deadline exceeded") && try < 2 {
					time.Sleep(time.Second)
					continue
				}

				if util.RegexCheck(tc.errRegex, e) {
					return
				}
				t.Fatalf("attempt %d: error mismatch\ngot:  %q\nwant: %q", try, e, tc.errRegex)
			}
		})
	}
}

func getPlain(client *http.Client, url string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequest() failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"status code mismatch\ngot:  %d - %q\nwant: %d",
			resp.StatusCode, body, http.StatusOK,
		)
	}
	return err
}
