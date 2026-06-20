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
	// GIVEN: a HTTP client, and a url to query.
	tests := []struct {
		name     string
		client   *http.Client
		urlKey   string
		errRegex string
	}{
		{
			name:     "Client accepts valid certificate",
			client:   Client,
			urlKey:   "url_valid",
			errRegex: `^$`,
		},
		{
			name:     "Client rejects invalid certificate",
			client:   Client,
			urlKey:   "url_invalid",
			errRegex: `x509`,
		},
		{
			name:     "InsecureClient accepts valid certificate",
			client:   InsecureClient,
			urlKey:   "url_valid",
			errRegex: `^$`,
		},
		{
			name:     "InsecureClient accepts invalid certificate",
			client:   InsecureClient,
			urlKey:   "url_invalid",
			errRegex: `^$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url := test.LookupPlain[tc.urlKey]

			try := 0
			for {
				try++

				// WHEN: a HTTP request is made to this URL.
				err := getPlain(t, tc.client, url)
				e := errfmt.FormatError(err)
				if strings.Contains(e, "context deadline exceeded") && try < 3 {
					time.Sleep(time.Second)
					continue
				}

				// THEN: an error is returned only when expected.
				if util.RegexCheck(tc.errRegex, e) {
					return
				}
				t.Fatalf(
					"%s\nGET %q error mismatch\ngot:  %q\nwant: %q",
					packageName, url, e, tc.errRegex,
				)
			}
		})
	}
}

func getPlain(t *testing.T, client *http.Client, url string) error {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	_, err = io.Copy(io.Discard, resp.Body)
	return err
}
