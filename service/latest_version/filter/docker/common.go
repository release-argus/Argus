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

package docker

import (
	"fmt"
	"time"

	"github.com/release-argus/Argus/internal/httpx"
)

// ########################
// # REGISTRY | UTILITIES #
// ########################

// check will query the registry for the image:tag of this version.
func check(version string, registry Registry) error {
	tag := registry.GetTagForVersion(version)

	// HTTP Request.
	req, err := registry.newRequest(tag)
	if err != nil {
		return err
	}

	// Auth.
	detail := registry.Detail()
	queryToken, err := registry.GetAuth().GetQueryToken(detail)
	if err != nil {
		return err //nolint:wrapcheck
	}
	if queryToken != "" {
		req.Header.Set("Authorization", "Bearer "+queryToken)
	}

	// Do the request.
	resp, err := httpx.Client.Do(req)
	if err != nil {
		return fmt.Errorf(
			"%s:%s %w",
			detail.Image, tag, err,
		)
	}

	// Parse the body.
	defer resp.Body.Close()
	// e.g. Quay will give a 200 even when the tag does not exist.
	if err := registry.parseBody(tag, resp); err != nil {
		return err
	}

	return nil
}

// ####################
// # AUTH | UTILITIES #
// ####################

// isUsable reports whether queryToken is non-empty and validUntil is at least two seconds in the future.
func isUsable(token string, validUntil time.Time) bool {
	return token != "" &&
		validUntil.After(time.Now().Add(2*time.Second).UTC())
}
