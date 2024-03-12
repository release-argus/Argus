// Copyright [2024] [Argus]
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

package shoutrrr

import (
	"encoding/json"
	"fmt"

	"github.com/release-argus/Argus/util"
)

// UseTemplate will create a new Shoutrrr object with the given overrides applied to this template.
func (s *Shoutrrr) UseTemplate(
	options *string,
	urlFields *string,
	params *string,

	logFrom *util.LogFrom,
) (shoutrrr *Shoutrrr, err error) {
	// Notify
	shoutrrr = &Shoutrrr{}
	shoutrrr.Type = s.Type

	err = stringToMap(options, &shoutrrr.Options, &s.Options)
	if err != nil {
		err = fmt.Errorf(`options:\  %s`, err)
		return
	}
	err = stringToMap(urlFields, &shoutrrr.URLFields, &s.URLFields)
	if err != nil {
		err = fmt.Errorf(`url_fields:\  %s`, err)
		return
	}
	err = stringToMap(params, &shoutrrr.Params, &s.Params)
	if err != nil {
		err = fmt.Errorf(`params:\  %s`, err)
		return
	}

	shoutrrr.Main = s.Main
	shoutrrr.Defaults = s.Defaults
	shoutrrr.HardDefaults = s.HardDefaults

	shoutrrr.ServiceStatus = s.ServiceStatus
	shoutrrr.ServiceStatus.Init(1, 0, 0, s.ServiceStatus.ServiceID, s.ServiceStatus.WebURL)
	shoutrrr.Failed = &shoutrrr.ServiceStatus.Fails.Shoutrrr

	err = shoutrrr.CheckValues("")
	return
}

// stringToMap will put baseMap into targetMap and convert str into a map[string]string and put it into targetMap.
//
// values of "<secret>" in str will be kept at the value in baseMap
func stringToMap(str *string, targetMap *map[string]string, baseMap *map[string]string) (err error) {
	if str == nil {
		*targetMap = *baseMap
		return
	}
	*targetMap = util.CopyMap(*baseMap)

	strMap := make(map[string]string)
	err = json.Unmarshal([]byte(*str), &strMap)
	if err != nil {
		return
	}

	for k, v := range strMap {
		if v != "<secret>" || (*targetMap)[k] == "" {
			(*targetMap)[k] = v
		}
	}
	return
}
