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

package gotify

import (
	"fmt"
	"strconv"
	"time"

	"github.com/release-argus/Argus/utils"
)

// CheckValues of this Slice.
func (g *Slice) CheckValues(prefix string) (errs error) {
	if g == nil {
		return
	}

	for key := range *g {
		if err := (*g)[key].CheckValues(prefix + "    "); err != nil {
			errs = fmt.Errorf("%s%s  %s:\\%w", utils.ErrorToString(errs), prefix, key, err)
		}
	}

	if errs != nil {
		errs = fmt.Errorf("%sgotify:\\%w", prefix, errs)
	}
	return
}

// CheckValues of this Gotification.
func (g *Gotify) CheckValues(prefix string) (errs error) {
	// Delay
	if g.Delay != nil {
		// Default to seconds when an integer is provided
		if _, err := strconv.Atoi(*g.Delay); err == nil {
			*g.Delay += "s"
		}
		if _, err := time.ParseDuration(*g.Delay); err != nil {
			errs = fmt.Errorf("%s%sdelay: <invalid> %q (Use 'AhBmCs' duration format)\\", utils.ErrorToString(errs), prefix, *g.Delay)
		}
	}

	if g.Main != nil {
		if g.GetURL() == nil {
			errs = fmt.Errorf("%s%surl: <required> (here or in gotify.%s)\\", utils.ErrorToString(errs), prefix, *g.ID)
		}
		if g.GetToken() == nil {
			errs = fmt.Errorf("%s%stoken: <required> (here or in gotify.%s)\\", utils.ErrorToString(errs), prefix, *g.ID)
		}
	}
	return
}

// Print the Slice.
func (g *Slice) Print(prefix string) {
	if g == nil {
		return
	}

	fmt.Printf("%sgotify:\n", prefix)
	for gotifyID, gotify := range *g {
		fmt.Printf("%s  %s:\n", prefix, gotifyID)
		gotify.Print(prefix + "    ")
	}
}

// Print the Gotify Struct.
func (g *Gotify) Print(prefix string) {
	utils.PrintlnIfNotNil(g.URL, fmt.Sprintf("%surl: %s", prefix, utils.DefaultIfNil(g.URL)))
	utils.PrintlnIfNotNil(g.Token, fmt.Sprintf("%stoken: %q", prefix, utils.DefaultIfNil(g.Token)))
	utils.PrintlnIfNotNil(g.Title, fmt.Sprintf("%stitle: %q", prefix, utils.DefaultIfNil(g.Title)))
	utils.PrintlnIfNotNil(g.Message, fmt.Sprintf("%smessage: %q", prefix, utils.DefaultIfNil(g.Message)))
	if g.Extras != nil &&
		(g.Extras.AndroidAction != nil ||
			g.Extras.ClientDisplay != nil ||
			g.Extras.ClientNotification != nil) {
		fmt.Printf("%sextras:\n", prefix)
		if g.Extras.AndroidAction != nil {
			utils.PrintlnIfNotNil(g.Extras.AndroidAction, fmt.Sprintf("%s  android_action: %q", prefix, utils.DefaultIfNil(g.Extras.AndroidAction)))
		}
		if g.Extras.ClientDisplay != nil {
			utils.PrintlnIfNotNil(g.Extras.ClientDisplay, fmt.Sprintf("%s  client_display: %q", prefix, utils.DefaultIfNil(g.Extras.ClientDisplay)))
		}
		if g.Extras.ClientNotification != nil {
			utils.PrintlnIfNotNil(g.Extras.ClientNotification, fmt.Sprintf("%s  client_notification: %q", prefix, utils.DefaultIfNil(g.Extras.ClientNotification)))
		}
	}
	utils.PrintlnIfNotNil(g.Priority, fmt.Sprintf("%spriority: %d", prefix, utils.DefaultIfNil(g.Priority)))
	utils.PrintlnIfNotNil(g.Delay, fmt.Sprintf("%sdelay: %s", prefix, utils.DefaultIfNil(g.Delay)))
	utils.PrintlnIfNotNil(g.MaxTries, fmt.Sprintf("%smax_tries: %d", prefix, utils.DefaultIfNil(g.MaxTries)))
}
