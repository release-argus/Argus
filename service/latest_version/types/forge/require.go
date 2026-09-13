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

// Package forge provides the release handling shared by git forge lookup types.
package forge

import (
	"fmt"
	"time"

	"github.com/release-argus/Argus/internal/logx"
	"github.com/release-argus/Argus/service/latest_version/filter"
	forgetypes "github.com/release-argus/Argus/service/latest_version/types/forge/api_type"
)

// ReleaseMeetsRequirements verifies that the `release` meets `require` for the `serviceID` service,
// and returns the version and its release date if it does.
func ReleaseMeetsRequirements(
	release forgetypes.Release,
	require *filter.Require,
	serviceID string,
	logFrom logx.LogFrom,
) (string, string, error) {
	version := release.TagName
	if release.SemanticVersion != nil {
		version = release.SemanticVersion.String()
	}
	releaseDate := release.PublishedAt

	// Verify the date is in RFC3339 format.
	if _, err := time.Parse(time.RFC3339, releaseDate); err != nil {
		logx.Warn(
			fmt.Errorf(
				"ignoring release date of %q for version %q on %q as it's not in RFC3339 format: %w",
				releaseDate, version, serviceID, err,
			),
			logFrom, releaseDate != "",
		)
		releaseDate = ""
	}

	if require == nil {
		return version, releaseDate, nil
	}

	// Version RegEx.
	if err := require.RegexCheckVersion(version, logFrom); err != nil {
		return "", "", err //nolint:wrapcheck
	}

	// Content RegEx.
	if assetReleaseDate, err := require.RegexCheckContentForge(version, release.Assets, logFrom); err != nil {
		return "", "", err //nolint:wrapcheck
	} else if assetReleaseDate != "" {
		releaseDate = assetReleaseDate
	}

	// Command.
	if err := require.ExecCommand(version, logFrom); err != nil {
		return "", "", err //nolint:wrapcheck
	}

	// Docker tag.
	if err := require.DockerTagCheck(version); err != nil {
		logx.Warn(err, logFrom, true)
		return "", "", err //nolint:wrapcheck
	} else if require.Docker != nil {
		logx.Info(
			fmt.Sprintf(
				`found %s container "%s:%s"`,
				require.Docker.GetType(), require.Docker.GetImage(), require.Docker.GetTagForVersion(version),
			),
			logFrom,
			true,
		)
	}

	return version, releaseDate, nil
}
