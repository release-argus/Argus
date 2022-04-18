// Copyright [2022] [Hymenaios]
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

package slack

import (
	"fmt"
	"strconv"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
)

// CheckValues of the Slice.
func (s *Slice) CheckValues(prefix string) (errs error) {
	if s == nil {
		return
	}

	for key := range *s {
		if err := (*s)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w", utils.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%sslack:\\%s", prefix, utils.ErrorToString(errs))
	}
	return
}

// CheckValues are valid.
func (s *Slack) CheckValues(prefix string) (errs error) {
	// Delay
	if s.Delay != nil {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(*s.Delay); err == nil {
			*s.Delay += "s"
		}
		if _, err := time.ParseDuration(*s.Delay); err != nil {
			errs = fmt.Errorf("%s%sdelay: <invalid> %q (Use 'AhBmCs' duration format)\\", utils.ErrorToString(errs), prefix, *s.Delay)
		}
	}

	if s.Main != nil {
		if s.GetURL() == nil {
			errs = fmt.Errorf("%s%surl: <required> (here or in slack.%s)\\", utils.ErrorToString(errs), prefix, *s.ID)
		}
	}
	return
}

// Print the Slice.
func (s *Slice) Print(prefix string) {
	if s == nil {
		return
	}

	fmt.Printf("%sslack:\n", prefix)
	for slackID, slack := range *s {
		fmt.Printf("%s  %s:\n", prefix, slackID)
		slack.Print(prefix + "    ")
	}
}

// Print The Slack struct.
func (s *Slack) Print(prefix string) {
	utils.PrintlnIfNotNil(s.URL, fmt.Sprintf("%surl: %q", prefix, utils.DefaultIfNil(s.URL)))
	utils.PrintlnIfNotNil(s.IconEmoji, fmt.Sprintf("%sicon_emoji: %q", prefix, utils.DefaultIfNil(s.IconEmoji)))
	utils.PrintlnIfNotNil(s.IconURL, fmt.Sprintf("%sicon_url: %q", prefix, utils.DefaultIfNil(s.IconURL)))
	utils.PrintlnIfNotNil(s.Username, fmt.Sprintf("%susername: %q", prefix, utils.DefaultIfNil(s.Username)))
	utils.PrintlnIfNotNil(s.Message, fmt.Sprintf("%smessage: %q", prefix, utils.DefaultIfNil(s.Message)))
	utils.PrintlnIfNotNil(s.Delay, fmt.Sprintf("%sdelay: %s", prefix, utils.DefaultIfNil(s.Delay)))
	utils.PrintlnIfNotNil(s.MaxTries, fmt.Sprintf("%smax_tries: %d", prefix, utils.DefaultIfNil(s.MaxTries)))
}
