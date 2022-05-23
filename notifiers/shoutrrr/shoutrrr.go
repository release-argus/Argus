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

package shoutrrr

import (
	"fmt"
	"strings"
	"time"

	shoutrrr_lib "github.com/containrrr/shoutrrr"
	shoutrrr_types "github.com/containrrr/shoutrrr/pkg/types"
	"github.com/release-argus/Argus/utils"
	"github.com/release-argus/Argus/web/metrics"
)

func (s *Shoutrrr) GetParams() (params *shoutrrr_types.Params) {
	p := make(shoutrrr_types.Params)
	params = &p

	// Service Params
	for key := range *s.Params {
		(*params)[key] = s.GetSelfParam(key)
	}

	// Main Params
	for key := range *s.Main.Params {
		_, exist := (*s.Main.Params)[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.Main.GetSelfParam(key)
		}
	}

	// Default Params
	for key := range *s.Defaults.Params {
		_, exist := (*s.Defaults.Params)[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.Defaults.GetSelfParam(key)
		}
	}

	// HardDefault Params
	for key := range *s.HardDefaults.Params {
		_, exist := (*s.HardDefaults.Params)[key]
		// Only overwrite if it doesn't exist in the level below
		if !exist {
			(*params)[key] = s.HardDefaults.GetSelfParam(key)
		}
	}

	return
}

func (s *Shoutrrr) GetURL() (url string) {
	switch s.Type {
	case "discord":
		// discord://token@webhookid
		url = fmt.Sprintf("discord://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("webhookid"))
	case "email":
		// smtp://username:password@host:port[/path]
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("smtp://%s:%s@%s:%s%s",
			s.GetURLField("username"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path))
	case "gotify":
		// gotify://host:port/path/token
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("gotify://%s%s%s/%s",
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("token"))
	case "googlechat":
		url = s.GetURLField("raw")
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "googlechat://")
		// googlechat://url
		url = fmt.Sprintf("googlechat://%s", url)
	case "ifttt":
		// ifttt://webhookid
		url = "ifttt://" + s.GetURLField("webhookid")
	case "join":
		// join://apiKey@join
		url = fmt.Sprintf("join://%s",
			s.GetURLField("apikey"))
	case "mattermost":
		// mattermost://[username@]host[:port][/path]/token[/channel]
		username := s.GetURLField("username")
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		channel := s.GetURLField("channel")
		url = fmt.Sprintf("mattermost://%s%s%s%s/%s%s",
			utils.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("token"),
			utils.ValueIfNotDefault(channel, "/"+channel))
	case "matrix":
		// matrix://user:password@host
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("matrix://%s%s@%s%s%s",
			s.GetURLField("user"),
			s.GetURLField("password"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path))
	case "opsgenie":
		// opsgenie://host[:port][/path]/apikey
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("opsgenie://%s%s%s/%s",
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("apikey"))
	case "pushbullet":
		// pushbullet://token/targets
		url = fmt.Sprintf("pushbullet://%s/%s",
			s.GetURLField("token"),
			s.GetURLField("targets"))
	case "pushover":
		// pushover://token@user
		url = fmt.Sprintf("pushover://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("user"))
	case "rocketchat":
		// rocketchat://[username@]host:port[/port]/tokenA/tokenB/channel
		username := s.GetURLField("username")
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("rocketchat://%s%s%s%s/%s/%s/%s",
			utils.ValueIfNotDefault(username, username+"@"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path),
			s.GetURLField("tokena"),
			s.GetURLField("tokena"),
			s.GetURLField("channel"))
	case "slack":
		// slack://token:token@channel
		url = fmt.Sprintf("slack://%s@%s",
			s.GetURLField("token"),
			s.GetURLField("channel"))
	case "teams":
		// teams://[group@][tenant][/altid][/groupowner]
		group := s.GetURLField("group")
		altid := strings.TrimPrefix(s.GetURLField("altid"), "/")
		groupowner := strings.TrimPrefix(s.GetURLField("groupowner"), "/")
		url = fmt.Sprintf("teams://%s%s%s%s",
			utils.ValueIfNotDefault(group, group+"@"),
			s.GetURLField("tenant"),
			utils.ValueIfNotDefault(altid, "/"+altid),
			utils.ValueIfNotDefault(groupowner, "/"+groupowner))
		url = strings.Replace(url, "///", "//", 1)
	case "telegram":
		// telegram://token@telegram
		url = fmt.Sprintf("telegram://%s@telegram",
			s.GetURLField("token"))
	case "zulipchat":
		// zulip://botMail:botKey@host:port
		port := strings.TrimPrefix(s.GetURLField("port"), ":")
		path := strings.TrimPrefix(s.GetURLField("path"), "/")
		url = fmt.Sprintf("zulip://%s:%s@%s%s%s",
			s.GetURLField("botmail"),
			s.GetURLField("botkey"),
			s.GetURLField("host"),
			utils.ValueIfNotDefault(port, ":"+port),
			utils.ValueIfNotDefault(path, "/"+path))
	case "shoutrrr":
		// Raw
		url = s.GetURLField("raw")
	}
	return
}

func (s *Slice) Send(
	title string,
	message string,
	serviceInfo *utils.ServiceInfo,
) error {
	if s == nil {
		return nil
	}
	if serviceInfo == nil {
		serviceInfo = &utils.ServiceInfo{}
	}

	errs := make(chan error)
	for key := range *s {
		// Send each message up to s.MaxTries number of times until they don't err.
		go func(shoutrrr *Shoutrrr) {
			logFrom := utils.LogFrom{Primary: *shoutrrr.ID, Secondary: serviceInfo.ID} // For logging
			triesLeft := shoutrrr.GetMaxTries()                                        // Number of times to send Shoutrrr (until 200 received).

			// Delay sending the Shoutrrr message by the defined interval.
			msg := fmt.Sprintf("%s, Sleeping for %s before sending the Shoutrrr message", *shoutrrr.ID, shoutrrr.GetDelay())
			jLog.Info(msg, logFrom, shoutrrr.GetDelay() != "0s")
			time.Sleep(shoutrrr.GetDelayDuration())

			url := shoutrrr.GetURL()
			sender, err := shoutrrr_lib.CreateSender(url)
			if err != nil {
				errs <- err
				return
			}
			params := shoutrrr.GetParams()
			if params == nil {
				p := make(shoutrrr_types.Params)
				params = &p
			}
			if title != "" {
				(*params)["title"] = title
			}
			toSend := message
			if message == "" {
				toSend = shoutrrr.GetMessage(serviceInfo)
			}
			combinedErrs := make(map[string]int)
			for {
				msg := fmt.Sprintf("Sending %q to %q", toSend, url)
				jLog.Verbose(msg, logFrom, !jLog.IsLevel("debug"))
				jLog.Debug(msg+fmt.Sprintf(" with params=%q", *params), logFrom, !jLog.IsLevel("debug"))
				err := sender.Send(toSend, params)

				failed := false
				for i := range err {
					if err[i] != nil {
						failed = true
						break
					}
				}

				// SUCCESS!
				if !failed {
					metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *shoutrrr.ID, serviceInfo.ID, shoutrrr.GetType(), "SUCCESS")
					failed := false
					shoutrrr.Failed = &failed
					errs <- nil
					return
				}

				// FAIL
				for new := range err {
					jLog.Error(err[new].Error(), logFrom, true)

					combinedErrs[err[new].Error()]++
					// errs = fmt.Errorf("%s%s  host: <required> e.g. 'mattermost.example.io'\\", utils.ErrorToString(errs), prefix)
				}
				metrics.InitPrometheusCounterActions(metrics.NotifyMetric, *shoutrrr.ID, serviceInfo.ID, shoutrrr.GetType(), "FAIL")
				triesLeft--

				// Give up after MaxTries.
				if triesLeft == 0 {
					msg = fmt.Sprintf("failed %d times to send a %s message to %s", shoutrrr.GetMaxTries(), shoutrrr.GetType(), shoutrrr.GetURL())
					jLog.Error(msg, logFrom, true)
					failed := true
					shoutrrr.Failed = &failed
					var err error
					for key := range combinedErrs {
						err = fmt.Errorf("%s%s x %d\n", utils.ErrorToString(err), key, combinedErrs[key])
					}
					errs <- err
					return
				}

				// Space out retries.
				time.Sleep(10 * time.Second)
			}
		}((*s)[key])
		// Space out Shoutrrr messages.const.
		time.Sleep(3 * time.Second)
	}

	var err error
	for range *s {
		errFound := <-errs
		if errFound != nil {
			if err == nil {
				err = errFound
			} else {
				err = fmt.Errorf("%s\\%s", err.Error(), errFound.Error())
			}
		}
	}
	return err
}
